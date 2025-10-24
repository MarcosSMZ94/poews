package utils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
)

var tradeRegex = regexp.MustCompile(`@(From) ([^:]+): (.+)`)

const (
	regexGroupCount    = 4
	formattedMsgSpaces = 4
)

var (
	ErrEmptyFile = errors.New("file is empty")
	ErrNoMatch   = errors.New("no regex match found")
)

func GetCurrentFileSize(filepath string) (int64, error) {
	stat, err := os.Stat(filepath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}
	return stat.Size(), nil
}

func GetNewTradeMessages(filepath string, lastOffset int64) ([]string, int64, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to stat file: %w", err)
	}

	fileSize := stat.Size()
	if fileSize == 0 {
		return nil, 0, ErrEmptyFile
	}

	if lastOffset >= fileSize {
		return nil, lastOffset, nil
	}

	_, err = file.Seek(lastOffset, 0)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to seek: %w", err)
	}

	var messages []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if message := MatchTradeMessage(line); message != "" {
			messages = append(messages, message)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, lastOffset, fmt.Errorf("scanner error: %w", err)
	}

	return messages, fileSize, nil
}

func MatchTradeMessage(line string) string {
	matches := tradeRegex.FindStringSubmatch(line)

	if len(matches) < regexGroupCount {
		return ""
	}

	from := matches[1]
	username := matches[2]
	message := matches[3]

	result := make([]byte, 0, len(from)+len(username)+len(message)+formattedMsgSpaces)
	result = append(result, from...)
	result = append(result, ' ')
	result = append(result, username...)
	result = append(result, ':', ' ')
	result = append(result, message...)

	return string(result)
}
