package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

type hexagramEntry struct {
	Name  string            `json:"卦名"`
	GuaCi string            `json:"卦辞"`
	YaoCi map[string]string `json:"爻辞"`
}

type hexagramExplainData struct {
	Name      string            `json:"name"`
	Upper     string            `json:"upper"`
	Lower     string            `json:"lower"`
	GuaCi     string            `json:"guaCi"`
	YaoCi     map[string]string `json:"yaoCi,omitempty"`
	RawKey    string            `json:"rawKey"`
	FoundBy   string            `json:"foundBy"`
	ErrorHint string            `json:"errorHint,omitempty"`
}

var (
	hexagramOnce sync.Once
	hexagramData map[string]hexagramEntry
	hexagramErr  error
)

func loadHexagramData() (map[string]hexagramEntry, error) {
	hexagramOnce.Do(func() {
		content, err := os.ReadFile("./conf/zhouyi.json")
		if err != nil {
			hexagramErr = err
			return
		}
		var parsed map[string]hexagramEntry
		if err := json.Unmarshal(content, &parsed); err != nil {
			hexagramErr = err
			return
		}
		hexagramData = parsed
	})
	return hexagramData, hexagramErr
}

// HexagramExplainHandler 查询卦象解释
func HexagramExplainHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	if r.Method != http.MethodGet {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("请求方法 %s 不支持", r.Method)
		writeJSON(w, res)
		return
	}

	name := strings.TrimSpace(r.URL.Query().Get("name"))
	upper := strings.TrimSpace(r.URL.Query().Get("upper"))
	lower := strings.TrimSpace(r.URL.Query().Get("lower"))

	if len(name) > 30 || len(upper) > 4 || len(lower) > 4 {
		res.Code = -1
		res.ErrorMsg = "参数长度不合法"
		writeJSON(w, res)
		return
	}

	data, err := loadHexagramData()
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "解释数据加载失败"
		writeJSON(w, res)
		return
	}

	entry, meta, err := findHexagramEntry(data, name, upper, lower)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = err.Error()
		writeJSON(w, res)
		return
	}

	res.Data = hexagramExplainData{
		Name:    entry.Name,
		Upper:   meta.upper,
		Lower:   meta.lower,
		GuaCi:   entry.GuaCi,
		YaoCi:   entry.YaoCi,
		RawKey:  meta.rawKey,
		FoundBy: meta.foundBy,
	}
	writeJSON(w, res)
}

type hexagramMeta struct {
	rawKey  string
	upper   string
	lower   string
	foundBy string
}

func findHexagramEntry(data map[string]hexagramEntry, name, upper, lower string) (hexagramEntry, hexagramMeta, error) {
	if upper != "" && lower != "" {
		key := fmt.Sprintf("%s上%s下", upper, lower)
		if entry, ok := data[key]; ok {
			return entry, hexagramMeta{rawKey: key, upper: upper, lower: lower, foundBy: "upperLower"}, nil
		}
	}

	cleanName := normalizeName(name)
	if cleanName != "" {
		for key, entry := range data {
			if normalizeName(entry.Name) == cleanName {
				upperGuess, lowerGuess := splitKey(key)
				return entry, hexagramMeta{rawKey: key, upper: upperGuess, lower: lowerGuess, foundBy: "name"}, nil
			}
		}
	}

	return hexagramEntry{}, hexagramMeta{}, errors.New("未找到卦象解释")
}

func normalizeName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "卦", "")
	return value
}

func splitKey(key string) (string, string) {
	parts := strings.Split(key, "上")
	if len(parts) != 2 {
		return "", ""
	}
	upper := parts[0]
	lower := strings.TrimSuffix(parts[1], "下")
	return upper, lower
}

func writeJSON(w http.ResponseWriter, res *JsonResult) {
	msg, err := json.Marshal(res)
	if err != nil {
		fmt.Fprint(w, "内部错误")
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(msg)
}
