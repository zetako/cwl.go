package client

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

// Timestamp 封装time.Time的数据结构
type Timestamp struct {
	time.Time
}

// String 格式化Timestamp
func (t Timestamp) String() string {
	return t.Format(time.RFC3339)
}

// Value 将Timestamp转换为int64
func (t Timestamp) Value() (driver.Value, error) {
	return t.Unix(), nil
}

// Scan 将int64转换为Timestamp
func (t *Timestamp) Scan(value interface{}) error {
	val, ok := value.(int64)
	if ok {
		t.Time = time.Unix(val, 0)
		return nil
	} else {
		val, ok := value.([]uint8)
		if ok {
			s := string(val)
			d, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				// return errors.Wrap(err, code.DATA_TRANSFORM_FAILED)
				return err
			}
			t.Time = time.Unix(d, 0)
			return nil
		} else {
			// return errors.New("can not convert value to Timestamp", code.DATA_TRANSFORM_FAILED)
			return fmt.Errorf("can not convert value to Timestamp")
		}
	}
}
