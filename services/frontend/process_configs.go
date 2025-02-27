package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Berops/claudie/internal/manifest"
	"github.com/Berops/claudie/proto/pb"
	cbox "github.com/Berops/claudie/services/context-box/client"
	"github.com/rs/zerolog/log"

	"gopkg.in/yaml.v3"
)

// processConfigs processes configs concurrently. If an error occurs while
// the file is being processed it's skipped and continues with
// the next one until all are processed. Nothing is done with
// files for which an error occurred, they'll be skipped until
// either corrected or deleted.
func (s *server) processConfigs() error {
	files, err := os.ReadDir(s.manifestDir)
	if err != nil {
		return fmt.Errorf("failed to read dir %q: %w", s.manifestDir, err)
	}

	log.Info().Msgf("Found %d files in %v", len(files), s.manifestDir)

	log.Info().Msg("Retrieving configs from context-box")

	configs, err := cbox.GetAllConfigs(s.cBox)
	if err != nil {
		return fmt.Errorf("failed to retrieve configs from context-box: %w", err)
	}

	log.Info().Msgf("Found %d configs in database", len(configs.Configs))

	type data struct {
		name        string
		rawManifest []byte
		path        string
	}

	dataChan := make(chan *data, len(files))
	group := sync.WaitGroup{}

	for _, file := range files {
		group.Add(1)

		// Process each of the files concurrently
		// in a separate go-routine skipping over
		// file for which an error occurs.
		go func(entry os.DirEntry) {
			defer group.Done()

			path := filepath.Join(s.manifestDir, entry.Name())
			rawManifest, err := os.ReadFile(path)
			if err != nil {
				log.Error().Msgf("skipping over file %v due to error: %v", path, err)
				return
			}

			m := manifest.Manifest{}
			if err := yaml.Unmarshal(rawManifest, &m); err != nil {
				log.Error().Msgf("skipping over file %v due to error: %v", path, err)
				return
			}

			if err := m.Validate(); err != nil {
				log.Error().Msgf("skipping over file %v due to error: %v", path, err)
				return
			}

			dataChan <- &data{
				name:        m.Name,
				rawManifest: rawManifest,
				path:        path,
			}
		}(file)
	}

	go func() {
		group.Wait()
		close(dataChan)
	}()

	// Collect data from files with no error.
	for data := range dataChan {
		configs.Configs = remove(configs.Configs, data.name)

		_, err := cbox.SaveConfigFrontEnd(s.cBox, &pb.SaveConfigRequest{
			Config: &pb.Config{
				Name:     data.name,
				Manifest: string(data.rawManifest),
			},
		})

		if err != nil {
			log.Error().Msgf("skip saving config: %v due to error: %v", data.name, err)
			continue
		}

		log.Info().Msgf("File %s has been saved to the database", data.path)
	}

	for _, config := range configs.Configs {
		if _, ok := s.deletingConfigs.Load(config.Id); ok {
			continue
		}

		s.deletingConfigs.Store(config.Id, nil)

		go func(config *pb.Config) {
			log.Info().Msgf("Deleting config: %v", config.Id)

			if err := cbox.DeleteConfig(s.cBox, config.Id, pb.IdType_HASH); err != nil {
				log.Error().Msgf("Failed to the delete %s with id %s : %v", config.Name, config.Id, err)
			}
			s.deletingConfigs.Delete(config.Id)
		}(config)
	}

	log.Info().Msg("Processed all files")

	return nil
}

// remove deletes the config with the specified name from the slice.
// If not present the original slice is returned.
func remove(configs []*pb.Config, configName string) []*pb.Config {
	for index, config := range configs {
		if config.Name == configName {
			configs = append(configs[0:index], configs[index+1:]...)
			break
		}
	}

	return configs
}
