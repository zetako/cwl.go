package slex

import "starlight/common/conf"

type BaseDirMapping struct {
	BaseDir map[string]string `json:"base_dir" yaml:"base_dir"`
}

var (
	globalConfig *BaseDirMapping
)

func initConfig() error {
	if globalConfig == nil {
		globalConfig = &BaseDirMapping{}
		err := conf.Unmarshal("slex", &globalConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	if err := initConfig(); err != nil {
		panic(err)
	}
}
