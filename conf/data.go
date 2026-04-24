package confdata

import (
	"embed"
	"errors"
)

//go:embed zhouyi.json
var zhouyiFS embed.FS

func LoadZhouyi() ([]byte, error) {
	data, err := zhouyiFS.ReadFile("zhouyi.json")
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("zhouyi.json 内容为空")
	}
	return data, nil
}
