package internal

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Properties struct {
	Server Server     `yaml:"Server"`
	JWT    JWT        `yaml:"JWT"`
	DB     DBProps    `yaml:"DB"`
	APIs   []APIRoute `yaml:"APIs"`
}

type DBProps struct {
	Host     string `yaml:"Host"`
	User     string `yaml:"User"`
	Password string `yaml:"Password"`
	Port     uint   `yaml:"Port"`
	DBName   string `yaml:"DBName"`
}

type Server struct {
	Port uint `yaml:"Port"`
}

type JWT struct {
	Secret string `yaml:"Secret"`
}

type APIRoute struct {
	Context string               `yaml:"Context"`
	Routes  []APIRoutePermission `yaml:"Routes"`
	Host    string               `yaml:"Host"`
}

type APIRoutePermission struct {
	Route  string `yaml:"Route"`
	Role   string `yaml:"Role"`
	Method string `yaml:"Method"`
}

func ParseYml(filepath string) Properties {
	var props Properties
	yamlFile, err := os.ReadFile(filepath)
	if err != nil {
		log.Panic(err.Error())
	}
	err = yaml.Unmarshal(yamlFile, &props)
	if err != nil {
		log.Panic(err.Error())
	}
	return props
}
