package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func readRepoNames(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	firstLine = strings.TrimRight(firstLine, "\n")

	_, firstLine, _ = strings.Cut(firstLine, ",")

	return strings.Split(firstLine, ","), nil
}

func queryPolls(repo string) (string, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/l7mp/%s", repo)
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	res.Body.Close()

	re := regexp.MustCompile(`"pull_count":([0-9]+),`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return "", nil
	}

	return string(matches[1]), nil
}

func main() {
	fileName := "pull-stats.csv"

	repos, err := readRepoNames(fileName)
	if err != nil {
		log.Fatalf("Unable to read %s: %s", fileName, err.Error())
	}

	row := []string{time.Now().Format("2006-01-02")}
	for _, r := range repos {
		polls, err := queryPolls(r)
		if err != nil {
			log.Printf("Unable to query %s: %s", r, err.Error())
		}
		row = append(row, polls)
	}

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Unable to open %s: %s", fileName, err.Error())
	}
	defer f.Close()

	s := strings.Join(row, ",") + "\n"
	if _, err := f.WriteString(s); err != nil {
		log.Fatalf("Unable to write data: %s", err.Error())
	}
}
