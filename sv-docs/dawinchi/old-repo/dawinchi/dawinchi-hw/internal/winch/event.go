package winch

import (
	"bufio"
	"encoding/json"
	"io"
	"os"

	"github.com/pkg/errors"
)

const (
	defaultEventLogFilePath = "/var/log/winch.log"
)

func (f *Firmware) EventLogFilePath() string {
	return f.eventLog.filePath
}

type EventLog struct {
	ID     uint64                 `json:"id"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

func (f *Firmware) AddEventLog(method string, event map[string]interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.eventLog.filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModeAppend|0644)
	if err != nil {
		return errors.Wrap(err, "open log file")
	}
	defer file.Close()

	newLogID := f.eventLog.id + 1
	if len(event) == 0 {
		event = make(map[string]interface{})
	}
	event["event_id"] = newLogID

	log := EventLog{
		ID:     newLogID,
		Method: method,
		Params: event,
	}

	jsonLog, err := json.Marshal(log)
	if err != nil {
		return errors.Wrap(err, "marshal log")
	}

	_, err = file.WriteString(string(jsonLog) + "\n")
	if err != nil {
		return errors.Wrap(err, "write log file")
	}

	// update last id
	f.eventLog.id = log.ID

	return nil
}

func (f *Firmware) eventLogLoadLastID() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.eventLog.filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrap(err, "open log file")
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return errors.Wrap(err, "stat log file")
	}

	// make sure we don't seek over the beginning of the file
	// and we are few lines back from the end of the file
	if fileInfo.Size() == 0 {
		f.eventLog.id = 0
		return nil
	}

	var seekBack int64 = 1024
	if fileInfo.Size() < seekBack {
		seekBack = fileInfo.Size()
	}

	_, err = file.Seek(-seekBack, io.SeekEnd)
	if err != nil {
		return errors.Wrap(err, "seek log file")
	}

	scanner := bufio.NewScanner(file)
	line := ""

	for scanner.Scan() {
		line = scanner.Text()
	}

	var lastLog EventLog
	err = json.Unmarshal([]byte(line), &lastLog)
	if err != nil {
		return errors.Wrap(err, "unmarshal log")
	}

	f.eventLog.id = lastLog.ID
	return nil
}

func (f *Firmware) EventLogsFromID(eventID uint64) ([]EventLog, error) {
	file, err := os.Open(f.eventLog.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	eventLogs := []EventLog{}

	for scanner.Scan() {
		l := scanner.Text()

		event := EventLog{}
		err = json.Unmarshal([]byte(l), &event)
		if err != nil {
			return nil, err
		}

		if event.ID >= eventID {
			continue
		}

		eventLogs = append(eventLogs, event)
	}

	if scanner.Err() != nil {
		return eventLogs, scanner.Err()
	}

	return eventLogs, nil
}
