package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

type QnA struct {
	Question string
	Answer   string
}

func main() {
	csvFileName := flag.String("csv", "problems.csv", "A csv file for QNA in the format question, answer (if the question or the answer contain ',' then the whole question or answer should be enclosed in double quotes")
	timeLimit := flag.Int("time", 10, "Time limit for the quiz in seconds")
	flag.Parse()

	qna := make(chan QnA)

	go readCsv(*csvFileName, qna)
	startQuiz(qna, *timeLimit)
}

// jaspreet manesar
func getString(a any) string {
	return fmt.Sprintf("%v", a)
}

func manageTimer(ctx context.Context, cancel context.CancelFunc, timeOut chan bool, timeLimit int, correct *int, incorrect *int, total int) {
	// var timer int = 0

	// for timer != timeLimit {
	// 	timer++
	// 	time.Sleep(1000 * time.Millisecond)
	// }
	timer := time.NewTimer(time.Duration(timeLimit) * time.Second)

	select {
	case <-timer.C:
		// Time is up, cancel the context
		cancel()
	case <-ctx.Done():
		// If context is cancelled for any reason, stop
		return
	}
	// fmt.Printf("Inside manage timer, time's up!\n")
	// stopQuiz(*correct, *incorrect, total)
	// timeOut <- true
}

func startQuiz(qna chan QnA, timeLimit int) {
	var qNo int = 1
	var correct int = 0
	var incorrect int = 0
	var total int = 14
	ctx, cancel := context.WithCancel(context.Background())
	timeOut := make(chan bool)
	defer close(timeOut)

	go manageTimer(ctx, cancel, timeOut, timeLimit, &correct, &incorrect, total)
	for {
		select {
		case data, ok := <-qna:
			if !ok {
				stopQuiz(correct, incorrect, total)
				return
			}

			select {
			case <-ctx.Done():
				stopQuiz(correct, incorrect, total)
				return
			default:
				fmt.Printf("Q%d. %s\n", qNo, data.Question)
				qNo++
				answerChan := make(chan string)
				scanner := bufio.NewScanner(os.Stdin)

				go func() {
					scanner.Scan()
					err := scanner.Err()
					if err != nil {
						log.Fatal("Scanner error: ", err)
					}
					answerChan <- scanner.Text()
				}()

				select {
				case <-ctx.Done():
					stopQuiz(correct, incorrect, total)
					return

				case userAns := <-answerChan:
					if getString(data.Answer) == userAns {
						correct++
					} else {
						incorrect++
					}
				}
			}

		case <-timeOut:
			// stopQuiz(correct, incorrect, total)
			return
		}
	}
}

func stopQuiz(correct int, incorrect int, total int) {
	if correct+incorrect != total {
		fmt.Printf("Time's up! You answered %d out of %d questions correctly.\n", correct, total)
	} else {
		fmt.Printf("Quiz completed! You answered %d out of %d questions correctly.\n", correct, total)
	}
}

func readCsv(fileName string, qna chan QnA) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	defer close(qna)

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	data, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	for _, row := range data {
		if len(row) >= 2 {
			qna <- QnA{Question: row[0], Answer: row[1]}
		}
	}
}
