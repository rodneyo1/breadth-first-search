package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Room struct {
	name    string
	x, y    int
	isStart bool
	isEnd   bool
}

type Link struct {
	from     string
	to       string
	capacity int
}

func parseInput(filename string) (int, []Room, []Link, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var ants int
	var rooms []Room
	var links []Link
	var nextIsStart, nextIsEnd bool

	for scanner.Scan() {
		line := scanner.Text()
		
		if line == "" {
			continue
		}
		if line == "##start" {
			nextIsStart = true
			continue
		}
		if line == "##end" {
			nextIsEnd = true
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}

		if ants == 0 {
			ants, err = strconv.Atoi(line)
			if err != nil || ants <= 0 {
				return 0, nil, nil, fmt.Errorf("invalid number of ants")
			}
			continue
		}

		if strings.Contains(line, "-") {
			parts := strings.Split(line, "-")
			links = append(links, Link{from: parts[0], to: parts[1], capacity: 1})
		} else {
			parts := strings.Fields(line)
			if len(parts) != 3 {
				continue
			}
			x, _ := strconv.Atoi(parts[1])
			y, _ := strconv.Atoi(parts[2])
			rooms = append(rooms, Room{
				name:    parts[0],
				x:       x,
				y:       y,
				isStart: nextIsStart,
				isEnd:   nextIsEnd,
			})
			nextIsStart = false
			nextIsEnd = false
		}
	}

	return ants, rooms, links, nil
}

// Simplified BFS that just returns if a path exists and the path itself
func findPath(links []Link, start, end string) (bool, []string) {
	visited := make(map[string]bool)
	parent := make(map[string]string)
	queue := []string{start}
	visited[start] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == end {
			// Reconstruct path
			path := []string{end}
			for current != start {
				current = parent[current]
				path = append([]string{current}, path...)
			}
			return true, path
		}

		// Check all links for possible moves
		for i := range links {
			if links[i].capacity > 0 {
				var next string
				if links[i].from == current {
					next = links[i].to
				} else if links[i].to == current {
					next = links[i].from
				} else {
					continue
				}

				if !visited[next] {
					visited[next] = true
					parent[next] = current
					queue = append(queue, next)
				}
			}
		}
	}

	return false, nil
}

// Updates link capacities for a given path
func updateCapacities(links []Link, path []string) {
	for i := 0; i < len(path)-1; i++ {
		from := path[i]
		to := path[i+1]
		// Find and update the corresponding link
		for j := range links {
			if (links[j].from == from && links[j].to == to) ||
				(links[j].from == to && links[j].to == from) {
				links[j].capacity = 0
				break
			}
		}
	}
}

func moveAnts(paths [][]string, antCount int) {
    // Calculate path lengths
    pathLengths := make([]int, len(paths))
    for i := range paths {
        pathLengths[i] = len(paths[i]) - 1
    }

    // Assign ants to optimal paths
    antAssignments := make([]int, antCount+1) // antID -> pathIndex
    for ant := 1; ant <= antCount; ant++ {
        bestPath := 0
        bestCost := pathLengths[0] + countAntsInPath(antAssignments, 0)
        
        for pathIdx := 1; pathIdx < len(paths); pathIdx++ {
            currentCost := pathLengths[pathIdx] + countAntsInPath(antAssignments, pathIdx)
            if currentCost < bestCost {
                bestPath = pathIdx
                bestCost = currentCost
            }
        }
        antAssignments[ant] = bestPath
    }

    // Track active ants and their positions
    type antState struct {
        pathIdx  int
        position int
        finished bool
    }
    
    antStates := make(map[int]*antState)
    activeAnts := 0  // Ants that have started but not finished
    startedAnts := 0 // Total ants that have started
    finishedAnts := 0

    for finishedAnts < antCount {
        moves := []string{}
        
        // Start new ants if possible
        if startedAnts < antCount {
            canStartNew := true
            // Check if first room in the path is available for the next ant
            nextAnt := startedAnts + 1
            pathIdx := antAssignments[nextAnt]
            
            for _, state := range antStates {
                if !state.finished && state.pathIdx == pathIdx && state.position == 1 {
                    canStartNew = false
                    break
                }
            }
            
            if canStartNew {
                antStates[nextAnt] = &antState{
                    pathIdx:  pathIdx,
                    position: 1,
                    finished: false,
                }
                moves = append(moves, fmt.Sprintf("L%d-%s", nextAnt, paths[pathIdx][1]))
                startedAnts++
                activeAnts++
            }
        }
        
        // Move active ants forward
        for ant := 1; ant <= startedAnts; ant++ {
            state, exists := antStates[ant]
            if !exists || state.finished {
                continue
            }
            
            if state.position < len(paths[state.pathIdx])-1 {
                nextPos := state.position + 1
                canMove := true
                
                // Check if next position is occupied
                for otherAnt, otherState := range antStates {
                    if otherAnt != ant && !otherState.finished &&
                       otherState.pathIdx == state.pathIdx && 
                       otherState.position == nextPos {
                        canMove = false
                        break
                    }
                }
                
                if canMove {
                    state.position = nextPos
                    moves = append(moves, fmt.Sprintf("L%d-%s", 
                        ant, paths[state.pathIdx][nextPos]))
                    
                    // Check if ant reached the end
                    if nextPos == len(paths[state.pathIdx])-1 {
                        state.finished = true
                        activeAnts--
                        finishedAnts++
                    }
                }
            }
        }

        if len(moves) > 0 {
            fmt.Println(strings.Join(moves, " "))
        }
    }
}

func countAntsInPath(antAssignments []int, pathIdx int) int {
    count := 0
    for _, path := range antAssignments {
        if path == pathIdx {
            count++
        }
    }
    return count
}
func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run . <input_file>")
		return
	}

	// Parse input
	ants, rooms, links, err := parseInput(os.Args[1])
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	// Find start and end rooms
	var start, end string
	for _, room := range rooms {
		if room.isStart {
			start = room.name
		}
		if room.isEnd {
			end = room.name
		}
	}

	// Find all paths using Edmonds-Karp concept
	var paths [][]string
	for {
		found, path := findPath(links, start, end)
		if !found {
			break
		}
		paths = append(paths, path)
		updateCapacities(links, path)
	}

	// Print input as required
	fmt.Println(ants)
	for _, room := range rooms {
		if room.isStart {
			fmt.Println("##start")
		}
		if room.isEnd {
			fmt.Println("##end")
		}
		fmt.Printf("%s %d %d\n", room.name, room.x, room.y)
	}
	for _, link := range links {
		fmt.Printf("%s-%s\n", link.from, link.to)
	}
	fmt.Println()

	// Move the ants
	moveAnts(paths, ants)
}