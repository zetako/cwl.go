package slex

import "starlight/common/conf"

type StarlightExecutorConfig struct {
	BaseDir map[string]string `json:"base_dir"`
}

var globalConfig *StarlightExecutorConfig

func initConfig() error {
	if globalConfig == nil {
		globalConfig = &StarlightExecutorConfig{}
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
