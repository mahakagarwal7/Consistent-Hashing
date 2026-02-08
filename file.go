package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
)

//Consistent Hashing Without virtual nodes
type ConsistentHasher struct {
	nodes []string
	slots []uint64
	ringSize uint64
}

//Hash ring
func NewConsistentHasher(ringSize uint64) *ConsistentHasher {
	return &ConsistentHasher{ringSize:ringSize}
}

//Add 
func (ch *ConsistentHasher) AddNode(node string) error {
	if len(ch.nodes) >= int(ch.ringSize) {
		return fmt.Errorf("ringSize exceeded")
	}
	nh := hashItem(node, ch.ringSize)
	slotIndex, found := slices.BinarySearch(ch.slots, nh)
	if found {
		return fmt.Errorf("collision")
	}
	ch.slots = slices.Insert(ch.slots, slotIndex, nh)
	ch.nodes = slices.Insert(ch.nodes, slotIndex, node)
	return nil
}

//Find 
func (ch *ConsistentHasher) FindNodeFor(item string) string {
	if len(ch.nodes) == 0 {
		panic("no nodes")
	}
	ih := hashItem(item, ch.ringSize)
	slotIndex, _ := slices.BinarySearch(ch.slots, ih)
	if slotIndex == len(ch.slots) {
		slotIndex = 0
	}
	return ch.nodes[slotIndex]
}

//Consistent Hashing with virtual nodes
type ConsistentHasherV struct {
	nodes []string
	slots []uint64
	ringSize uint64
	vnodesPerNode int
}

//Default virtual node
func NewConsistentHasherV(ringSize uint64) *ConsistentHasherV {
	return &ConsistentHasherV{ringSize: ringSize, vnodesPerNode: 10}
}

//Custom virtual node
func NewConsistentHasherVWithVNodes(ringSize uint64, vnodesPerNode int) *ConsistentHasherV {
	return &ConsistentHasherV{ringSize: ringSize, vnodesPerNode: vnodesPerNode}
}

//Add
func (ch *ConsistentHasherV) AddNode(node string) error {
	if len(ch.nodes) >= int(ch.ringSize)-ch.vnodesPerNode {
		return fmt.Errorf("ringSize exceeded")
	}
	for i := range ch.vnodesPerNode {
		vnode := fmt.Sprintf("%v@%v", node, i)
		nh := hashItem(vnode, ch.ringSize)
		slotIndex, found := slices.BinarySearch(ch.slots, nh)
		if found {
			return fmt.Errorf("collision")
		}
		ch.slots = slices.Insert(ch.slots, slotIndex, nh)
		ch.nodes = slices.Insert(ch.nodes, slotIndex, vnode)
	}
	return nil
}


//Remove
func (ch *ConsistentHasherV) RemoveNode(node string) error {
	for i := range ch.vnodesPerNode {
		vnode := fmt.Sprintf("%v@%v", node, i)
		nh := hashItem(vnode, ch.ringSize)
		slotIndex, found := slices.BinarySearch(ch.slots, nh)
		if !found {
			return fmt.Errorf("vnode not found")
		}
		ch.slots = slices.Delete(ch.slots, slotIndex, slotIndex+1)
		ch.nodes = slices.Delete(ch.nodes, slotIndex, slotIndex+1)
	}
	return nil
}

//Find
func (ch *ConsistentHasherV) FindNodeFor(item string) string {
	if len(ch.nodes) == 0 {
		panic("no nodes")
	}
	ih := hashItem(item, ch.ringSize)
	slotIndex, _ := slices.BinarySearch(ch.slots, ih)
	if slotIndex == len(ch.slots) {
		slotIndex = 0
	}
	return nodeFromVnode(ch.nodes[slotIndex])
}


//Real Node
func nodeFromVnode(vnodeName string) string {
	node, _, found := strings.Cut(vnodeName, "@")
	if !found {
		panic("invalid vnode")
	}
	return node
}

//Hash a string 
func hashItem(item string, nslots uint64) uint64 {
	digest := md5.Sum([]byte(item))
	digestHigh := binary.BigEndian.Uint64(digest[8:16])
	digestLow := binary.BigEndian.Uint64(digest[:8])
	return (digestHigh ^ digestLow) % nslots
}


func generateRandomString(r *rand.Rand, length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.IntN(len(charset))]
	}
	return string(b)
}

type RebalanceMetrics struct {
	KeysMoved          int
	TotalKeys          int
	Percentage         float64
	ExpectedPercentage float64
}


func (rm *RebalanceMetrics) String() string {
	return fmt.Sprintf("Moved: %d/%d (%.2f%%), Expected: ~%.2f%%",
		rm.KeysMoved, rm.TotalKeys, rm.Percentage*100, rm.ExpectedPercentage*100)
}


type LoadDistribution struct {
	Mean      float64
	StdDev    float64
	MaxDevPct float64
}


func (ld *LoadDistribution) String() string {
	return fmt.Sprintf("Mean=%.1f, StdDev=%.2f, MaxDev=%.1f%%", ld.Mean, ld.StdDev, ld.MaxDevPct)
}



func analyzeDistribution(ch interface{ FindNodeFor(string) string }, keys []string, servers []string) *LoadDistribution {
	counts := make(map[string]int)
	for _, s := range servers {
		counts[s] = 0
	}
	for _, key := range keys {
		server := ch.FindNodeFor(key)
		counts[server]++
	}
	total := len(keys)
	mean := float64(total) / float64(len(servers))
	var sumSqDiff float64
	maxCount := 0
	for _, c := range counts {
		diff := float64(c) - mean
		sumSqDiff += diff * diff
		if c > maxCount {
			maxCount = c
		}
	}
	variance := sumSqDiff / float64(len(servers))
	return &LoadDistribution{
		Mean:      mean,
		StdDev:    math.Sqrt(variance),
		MaxDevPct: math.Abs(float64(maxCount)-mean) / mean * 100,
	}
}


func (ch *ConsistentHasherV) AddNodeWithMetrics(node string, testKeys []string) (*RebalanceMetrics, error) {
	before := make(map[string]string, len(testKeys))
	for _, key := range testKeys {
		before[key] = ch.FindNodeFor(key)
	}
	if err := ch.AddNode(node); err != nil {
		return nil, err
	}
	moved := 0
	for _, key := range testKeys {
		if ch.FindNodeFor(key) != before[key] {
			moved++
		}
	}
	return &RebalanceMetrics{
		KeysMoved:          moved,
		TotalKeys:          len(testKeys),
		Percentage:         float64(moved) / float64(len(testKeys)),
		ExpectedPercentage: 1.0 / float64(ch.getPhysicalNodeCount()),
	}, nil
}


func (ch *ConsistentHasherV) RemoveNodeWithMetrics(node string, testKeys []string) (*RebalanceMetrics, error) {
	before := make(map[string]string, len(testKeys))
	for _, key := range testKeys {
		before[key] = ch.FindNodeFor(key)
	}
	if err := ch.RemoveNode(node); err != nil {
		return nil, err
	}
	moved := 0
	for _, key := range testKeys {
		after := ch.FindNodeFor(key)
		if before[key] == node || after != before[key] {
			moved++
		}
	}
	return &RebalanceMetrics{
		KeysMoved:          moved,
		TotalKeys:          len(testKeys),
		Percentage:         float64(moved) / float64(len(testKeys)),
		ExpectedPercentage: 1.0 / float64(ch.getPhysicalNodeCount()),
	}, nil
}

func (ch *ConsistentHasherV) getPhysicalNodeCount() int {
	seen := make(map[string]bool)
	for _, vnode := range ch.nodes {
		seen[nodeFromVnode(vnode)] = true
	}
	return len(seen)
}


func main() {
	ringSize := uint64(1 << 32)
	servers := []string{"S0", "S1", "S2", "S3", "S4", "S5"}

	keys := make([]string, 1000000)
	rnd := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	for i := 0; i < 1000000; i++ {
		keys[i] = generateRandomString(rnd, 32)
	}

	fmt.Println("Without virtual nodes:")
	basic := NewConsistentHasher(ringSize)
	for _, s := range servers {
		basic.AddNode(s)
	}
	fmt.Println(analyzeDistribution(basic, keys, servers))

	fmt.Println("With virtual nodes (10 per server):")
	vnode := NewConsistentHasherV(ringSize)
	for _, s := range servers {
		vnode.AddNode(s)
	}
	fmt.Println(analyzeDistribution(vnode, keys, servers))

	fmt.Println("Adding S6:")
	metrics, _ := vnode.AddNodeWithMetrics("S6", keys)
	fmt.Println(metrics)

	fmt.Println("Removing S2:")
	removeMetrics, _ := vnode.RemoveNodeWithMetrics("S2", keys)
	fmt.Println(removeMetrics)

	fmt.Println("Vnode comparison (1,10,100,500):")
	for _, vn := range []int{1, 10, 100, 500} {
		ch := NewConsistentHasherVWithVNodes(ringSize, vn)
		for _, s := range servers {
			ch.AddNode(s)
		}
		d := analyzeDistribution(ch, keys, servers)
		fmt.Printf("VNodes=%d StdDev=%.2f\n", vn, d.StdDev)
	}
}



