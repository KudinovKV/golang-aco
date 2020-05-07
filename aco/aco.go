package aco

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/cheggaaa/pb/v3"
	zl "github.com/rs/zerolog/log"
	"gonum.org/v1/gonum/mat"
)

type Cluster struct {
	City      map[int][]Node
	CityCount int
	Vision    *mat.Dense
	Pheromone *mat.Dense
	mutex     *sync.Mutex
	dataCh    chan *Ant
	wg        *sync.WaitGroup
	workC     *sync.Cond
	isExit    bool
	BestPath  []int
	BestCost  float64
	newIter   *sync.Cond
}

type Node struct {
	X int
	Y int
}

type Ant struct {
	path         []int
	taby         []bool
	startCluster int
	length       float64
}

var (
	MAX_COST    = 1000000.0
	MAX_CLUSTER = 50
	MIN_CLUSTER = 40
	MAX_NODES   = 1002
	MIN_NODES   = 1000
	MAX_LENGTH  = 10000
	ALFA        = 1.0
	BETA        = 2.0
	Q           = float64(MIN_CLUSTER * MIN_NODES)
	R0          = 0.3
	bar         *pb.ProgressBar
	LIMIT       = 100
)

func NewAnt(startCluster, countCluster int) *Ant {
	t := make([]bool, countCluster)
	for i, _ := range t {
		t[i] = false
	}
	return &Ant{
		path:         make([]int, countCluster),
		taby:         t,
		startCluster: startCluster,
		length:       0.0,
	}
}

func NewCluster(count int) *Cluster {
	v := make([]float64, count*count)
	p := make([]float64, count*count)

	for i := 0; i != count*count; i++ {
		v[i] = MAX_COST
		p[i] = rand.Float64()
	}
	c := Cluster{
		City:      make(map[int][]Node),
		CityCount: count,
		Vision:    mat.NewDense(count, count, v),
		Pheromone: mat.NewDense(count, count, p),
		mutex:     &sync.Mutex{},
		dataCh:    make(chan *Ant),
		wg:        &sync.WaitGroup{},
		workC:     sync.NewCond(&sync.Mutex{}),
		isExit:    false,
		BestPath:  []int{},
		BestCost:  MAX_COST,
	}
	return &c
}

func GenerateNewCluster() *Cluster {
	zl.Debug().Msg("Start create new cluster")
	countCluster := rand.Intn(MAX_CLUSTER-MIN_CLUSTER) + MIN_CLUSTER
	countNodes := rand.Intn(MAX_NODES-MIN_NODES) + MIN_NODES
	newCluster := NewCluster(countNodes)
	for i := 0; i != countNodes; i++ {
		currentCluster := rand.Intn(countCluster)
		newCluster.City[currentCluster] = append(newCluster.City[currentCluster], Node{
			X: rand.Intn(MAX_LENGTH),
			Y: rand.Intn(MAX_LENGTH),
		})
	}
	zl.Debug().Msg("Finish create new cluster")
	return newCluster
}

func (c *Cluster) Distance(i, j int) (float64, error) {
	k := 0
	first := Node{}
	second := Node{}
	for _, cluster := range c.City {
		for _, node := range cluster {
			if k == i {
				first = node
			}
			if k == j {
				second = node
				return math.Sqrt(math.Pow(float64(first.X-second.X), 2) + math.Pow(float64(first.Y-second.Y), 2)), nil
			}
			k += 1
		}
	}
	return 0.0, errors.New("Can't find element")
}

func (c *Cluster) SetRandomArc(pos int) {
	i := 0
	for _, cluster := range c.City {
		for _, _ = range cluster {
			if i != pos {
				arc := rand.Float64()
				if arc > 0.5 {
					x, err := c.Distance(pos, i)
					if err != nil {
						zl.Fatal().Err(err)
					}
					c.Vision.Set(pos, i, x)
				}
			}
		}
	}
}

func (c *Cluster) PrintMatrix() {
	for i := 0; i != c.CityCount; i += 1 {
		for j := 0; j != c.CityCount; j += 1 {
			fmt.Print(c.Vision.At(i, j))
		}
		fmt.Println()
	}
}

func (c *Cluster) InitMatrix() {
	i := 0
	for _, cluster := range c.City {
		for _, _ = range cluster {
			c.SetRandomArc(i)
			i += 1
		}
	}
}

func (c *Cluster) onRoute(route []int, a, b int) bool {
	for i := 0; i != len(route)-1; i++ {
		if route[i] == a && route[i+1] == b {
			return true
		}
	}
	return false
}

func (c *Cluster) updatePheromone(ants []Ant) {
	delta := make([][]float64, c.CityCount)
	for i := 0; i != c.CityCount; i++ {
		delta[i] = make([]float64, c.CityCount)
		for j := 0; j != c.CityCount; j++ {
			delta[i][j] = 0.0
		}
	}
	for _, ant := range ants {
		for i := 0; i != c.CityCount; i++ {
			for j := 0; j != c.CityCount; j++ {
				if i == j {
					delta[i][j] += 0.0
				} else if c.onRoute(ant.path, i, j) == true {
					delta[i][j] += Q / ant.length
				} else {
					delta[i][j] += 0.0
				}
			}
		}
	}
	for i := 0; i != c.CityCount; i++ {
		for j := 0; j != c.CityCount; j++ {
			c.Pheromone.Set(i, j, (1-R0)*c.Pheromone.At(i, j)+delta[i][j])
		}
	}
}

func (c *Cluster) listener() {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	zl.Debug().Msg("Start listen")
	ants := []Ant{}
	counter := 0
	for ant := range c.dataCh {
		ants = append(ants, *ant)
		if ant.length < c.BestCost {
			counter = 0
			c.BestCost = ant.length
			for i := 0; i != len(ant.path); i++ {
				c.BestPath[i] = ant.path[i]
			}
		}
		if len(ants) == len(c.City) {
			c.workC.L.Lock()
			counter++
			if counter == LIMIT {
				c.isExit = true
			}
			c.updatePheromone(ants)
			ants = []Ant{}
			bar.Increment()
			c.workC.L.Unlock()
			c.workC.Broadcast()
		}
	}
}

func (c *Cluster) getStep(Ant *Ant, n int, numStep int) {
	p := make([]float64, c.CityCount)
	pAll := 0.0
	for i := 0; i != len(c.City); i++ {
		if Ant.taby[i] == true {
			continue
		}
		for j := 0; j != len(c.City[i]); j++ {
			pAll += math.Pow(c.Pheromone.At(n, j), ALFA) * math.Pow(c.Vision.At(n, j), BETA)
		}
	}
	k := 0
	for i := 0; i != len(c.City); i++ {
		for j := 0; j != len(c.City[i]); j++ {
			if Ant.taby[i] == true {
				p[k] = 0.0
			} else {
				p[k] = math.Pow(c.Pheromone.At(n, k), ALFA) * math.Pow(c.Vision.At(n, k), BETA) / pAll
			}
			k += 1
		}
	}
	maxP := 0.0
	maxNode := 0
	for i := 0; i != len(p); i++ {
		if maxP < p[i] {
			maxP = p[i]
			maxNode = i
		}
	}
	currentNode := Node{}
	k = 0
	for i := 0; i != len(c.City); i++ {
		for j := 0; j != len(c.City[i]); j++ {
			if k == n {
				currentNode = c.City[i][j]
			}
			k++
		}
	}
	k = 0
	for i := 0; i != len(c.City); i++ {
		for j := 0; j != len(c.City[i]); j++ {
			if k == maxNode {
				Ant.path[numStep] = k
				Ant.length += math.Sqrt(math.Pow(float64(currentNode.X-c.City[i][j].X), 2) + math.Pow(float64(currentNode.Y-c.City[i][j].Y), 2))
				Ant.taby[i] = true
				return
			}
			k++
		}
	}
	zl.Fatal().Msg("Something going wrong")
}

func (c *Cluster) createRoute(Ant *Ant) {
	for i := 1; i != len(c.City); i += 1 {
		c.getStep(Ant, Ant.path[i-1], i)
	}
	for _, taby := range Ant.taby {
		if taby != true {
			zl.Fatal().Msg("Visited not all city")
		}
	}
}

func (a *Ant) Clear() {
	for i, _ := range a.taby {
		a.taby[i] = false
		a.path[i] = 0
	}
	a.length = 0.0
}

func (c *Cluster) startWork(Ant *Ant) {
	defer c.wg.Done()
	for {
		if c.isExit == true {
			return
		}
		Ant.taby[Ant.startCluster] = true
		k := 0
		for i := 0; i != len(c.City); i++ {
			if Ant.startCluster == i {
				n := rand.Intn(len(c.City[i]))
				Ant.path[0] = k + n
				break
			} else {
				k += len(c.City[i])
			}
		}
		c.createRoute(Ant)

		//		c.workM.Lock()
		//		c.dataCh <- Ant
		//		c.workM.Unlock()

		c.workC.L.Lock()
		c.dataCh <- Ant
		c.workC.Wait()
		Ant.Clear()
		c.workC.L.Unlock()
	}
}

func ACO(clusters *Cluster) {
	zl.Debug().Msg("Start work")
	bar = pb.StartNew(1000)
	clusters.BestPath = make([]int, len(clusters.City))
	clusters.InitMatrix()
	go clusters.listener()
	for i := 0; i != len(clusters.City); i++ {
		clusters.wg.Add(1)
		newAnt := NewAnt(i, len(clusters.City))
		go clusters.startWork(newAnt)
	}
	clusters.wg.Wait()
	bar.Finish()
}
