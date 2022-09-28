package irunner

import (
  "fmt"
  "sync"
  "time"
)

//type Process struct {
//  ID string `json:"id"`
//  //Inputs []InputParameter `json:"inputs"`
//}

type MainLog struct {
  ProcessRequest
  *Log
}

type ProcessRequest struct {
  Process []byte   `json:"process"`
  Params    []byte   `json:"params"`
  //Configs    []byte   `json:"configs"`
  UserName   string            `json:"username"`
  Tags     map[string]string `json:"tags,omitempty"` // optional set of key:val pairs provided by user to annotate workflow run - NOTE: val is a string
  JobName  string            `json:"jobName,omitempty"` // populated internally by server
}

type Log struct {
  Created        string                 `json:"created,omitempty"` // timezone???
  CreatedObj     time.Time              `json:"-"`
  LastUpdated    string                 `json:"lastUpdated,omitempty"` // timezone???
  LastUpdatedObj time.Time              `json:"-"`
  JobID          string                 `json:"jobID,omitempty"`
  JobName        string                 `json:"jobName,omitempty"`
  ContainerImage string                 `json:"containerImage,omitempty"`
  Status         string                 `json:"status"`
  Stats          *Stats                 `json:"stats"`
  Event          *EventLog              `json:"eventLog,omitempty"`
  Input          map[string]interface{} `json:"input"`
  Output         map[string]interface{} `json:"output"`
  Scatter        map[int]*Log           `json:"scatter,omitempty"`
}

type Stats struct {
  CPUReq        ResourceRequirement `json:"cpuReq"` // in-progress
  MemoryReq     ResourceRequirement `json:"memReq"` // in-progress
  ResourceUsage ResourceUsage       `json:"resourceUsage"`
  Duration      float64             `json:"duration"`  // okay - currently measured in minutes
  DurationObj   time.Duration       `json:"-"`         // okay
  NFailures     int                 `json:"nfailures"` // TODO
  NRetries      int                 `json:"nretries"`  // TODO
}

// ResourceRequirement is for logging resource requests vs. actual usage
type ResourceRequirement struct {
  Min int64 `json:"min"`
  Max int64 `json:"max"`
}

// ResourceUsage ..
type ResourceUsage struct {
  Series         ResourceUsageSeries `json:"data"`
  SamplingPeriod int                 `json:"samplingPeriod"`
}

// ResourceUsageSeries ..
type ResourceUsageSeries []ResourceUsageSamplePoint

// ResourceUsageSamplePoint ..
type ResourceUsageSamplePoint struct {
  CPU    int64 `json:"cpu"`
  Memory int64 `json:"mem"`
}

type EventLog struct {
  sync.RWMutex
  Events []string `json:"events,omitempty"`
}

func logger() *Log {
  logger := &Log{
    Status: notStarted,
    Input:  make(map[string]interface{}),
    Stats:  &Stats{},
    Event:  &EventLog{},
  }
  logger.Event.info("init log")
  return logger
}

// a record is "<timestamp> - <level> - <message>"
func (log *EventLog) write(level, message string) {
  log.Lock()
  defer log.Unlock()
  timestamp := ts()
  record := fmt.Sprintf("%v - %v - %v", timestamp, level, message)
  log.Events = append(log.Events, record)
}

func (log *EventLog) infof(f string, v ...interface{}) {
  m := fmt.Sprintf(f, v...)
  log.info(m)
}

func (log *EventLog) info(m string) {
  log.write(infoLogLevel, m)
}

func (log *EventLog) warnf(f string, v ...interface{}) {
  m := fmt.Sprintf(f, v...)
  log.warn(m)
}

func (log *EventLog) warn(m string) {
  log.write(warningLogLevel, m)
}

func (log *EventLog) errorf(f string, v ...interface{}) error {
  m := fmt.Sprintf(f, v...)
  return log.error(m)
}

func (log *EventLog) error(m string) error {
  log.write(errorLogLevel, m)
  return fmt.Errorf(m)
}


func (e *Engine) errorf(f string, v ...interface{}) error {
  return e.Log.Event.errorf(f, v...)
}

func (l *Log) errorf(f string, v ...interface{}) error {
  return l.Event.errorf(f, v...)
}


func (l *Log) error(m string) error {
  return l.Event.error(m)
}

// get string timestamp for right now
func ts() string {
  t := time.Now()
  s := timef(t)
  return s
}

func timef(t time.Time) string {
  s := fmt.Sprintf("%v/%v/%v %v:%v:%v", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
  return s
}


