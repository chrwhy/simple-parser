package parser

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"
)

// SynonymDict 同义词词典
type SynonymDict struct {
	mu sync.RWMutex
	// forward: 主词 -> [同义词1, 同义词2, ...]
	forward map[string][]string
	// reverse: 同义词 -> [主词1, 主词2, ...]（用于反向查找）
	reverse map[string][]string
}

// LoadSynonymDict 从文件加载同义词词典。
// 文件格式：每行一个条目，格式为 "主词:同义词1,同义词2,..."
// 注释行以 # 开头，空行会被跳过。
func LoadSynonymDict(path string) (*SynonymDict, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseSynonymDict(f)
}

// parseSynonymDict 从 io.Reader 解析同义词词典。
func parseSynonymDict(r io.Reader) (*SynonymDict, error) {
	dict := &SynonymDict{
		forward: make(map[string][]string),
		reverse: make(map[string][]string),
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}

		synonyms := strings.Split(parts[1], ",")
		for _, syn := range synonyms {
			syn = strings.TrimSpace(syn)
			if syn == "" {
				continue
			}
			dict.forward[key] = append(dict.forward[key], syn)
			dict.reverse[syn] = append(dict.reverse[syn], key)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return dict, nil
}

// Lookup 查询一个词的所有同义词。
// 返回值包含该词作为主词时的所有同义词，以及该词作为同义词时的所有主词。
// 返回的同义词去重，不包含输入词本身。
func (d *SynonymDict) Lookup(word string) []string {
	if d == nil {
		return nil
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	seen := make(map[string]bool)
	var result []string

	// 查找作为主词时的同义词
	if syns, ok := d.forward[word]; ok {
		for _, syn := range syns {
			if syn != word && !seen[syn] {
				seen[syn] = true
				result = append(result, syn)
			}
		}
	}

	// 查找作为同义词时的主词
	if keys, ok := d.reverse[word]; ok {
		for _, key := range keys {
			if key != word && !seen[key] {
				seen[key] = true
				result = append(result, key)
			}
		}
		// 主词的其他同义词也加入
		for _, key := range keys {
			if syns, ok := d.forward[key]; ok {
				for _, syn := range syns {
					if syn != word && !seen[syn] {
						seen[syn] = true
						result = append(result, syn)
					}
				}
			}
		}
	}

	return result
}

// LookupPhrase 查询一个短语的同义词（整词匹配，不分词）。
// 用于短查询场景：直接将整个查询作为 key 在词典中查找。
func (d *SynonymDict) LookupPhrase(phrase string) []string {
	if d == nil {
		return nil
	}
	return d.Lookup(phrase)
}

// Size 返回词典中的条目数（主词数量）。
func (d *SynonymDict) Size() int {
	if d == nil {
		return 0
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.forward)
}
