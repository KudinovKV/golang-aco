package main

import (
	"strings"

	"github.com/KudinovKV/golang-aco/aco"
	"github.com/KudinovKV/golang-aco/config"
	"github.com/KudinovKV/golang-aco/filework"
	"github.com/KudinovKV/golang-aco/plot"
	zl "github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Config()
	if err != nil {
		zl.Fatal().Err(err).
			Msg("Can't read env")
	}

	var clusters *aco.Cluster
	if strings.EqualFold(cfg.Mode, "generate") {
		clusters = aco.GenerateNewCluster()
	} else if strings.EqualFold(cfg.Mode, "read") {
		clusters = filework.ReadInFile(cfg.InPath)
	} else {
		zl.Fatal().Err(err).
			Msg("Incorrect mode")
	}

	aco.ACO(clusters)
	zl.Debug().Msgf("Best cost : %v , best path : %v", clusters.BestCost, clusters.BestPath)
	plot.Draw(clusters)
}
