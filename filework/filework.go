package filework

import (
	"bufio"
	"github.com/KudinovKV/golang-aco/aco"
	zl "github.com/rs/zerolog/log"
	"log"
	"os"
	"strconv"
	"strings"
)

func ReadInFile(fileName string) *aco.Cluster {
	var clusters *aco.Cluster
	inFile, err := os.Open(fileName)
	if err != nil {
		zl.Fatal().Err(err).
			Msgf("Can't open in file %v" , fileName)
	}
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)
	i := 0
	for scanner.Scan() {
		if i == 3 {
			fields := strings.Fields(scanner.Text())
			currentNodes, err := strconv.Atoi(fields[1])
			if err != nil {
				zl.Fatal().Err(err).Msg("Can't read count")
			}
			clusters = aco.NewCluster(currentNodes)
		}
		if i >= 6 {
			if strings.EqualFold(scanner.Text() , "EOF"){
				break
			}
			log.Print(scanner.Text())
			fields := strings.Fields(scanner.Text())
			x , _ := strconv.Atoi(fields[0])
			y , _ := strconv.Atoi(fields[1])
			c , _ := strconv.Atoi(fields[2])
			log.Print(x , " " , y , " " , c)
			clusters.City[c] = append(clusters.City[c] , aco.Node{
				X:x,
				Y:y,
			})
		}
		i += 1
	}
	zl.Debug().
		Msg("Correctly read in file")
	return clusters
}