package datatable

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/c-robinson/iplib"
	"github.com/ca17/datahub/plugin/pkg/common"
	"github.com/ca17/datahub/plugin/pkg/httpc"
	"github.com/ca17/datahub/plugin/pkg/validutil"
)

const (
	WhichTypePath = iota
	WhichTypeUrl
	WhichTypeInline         // Dummy
	DateTypeKeywordTable    = "keyword_table"
	DateTypeNetlistTable    = "netlist_table"
	DateTypeDomainlistTable = "domain_table"
	DateTypeEcsTable        = "ecs_table"
)

type TextData interface {
	ParseFile(reader io.Reader) error
	ParseLines(lines []string, reset bool)
	ParseInline(ws []string)
	Match(name string) bool
	MatchNet(inet iplib.Net) bool
	LessString() string
	Reset()
	ForEach(f func(interface{}) error, mx int)
	Len() int
}

type DataTable struct {
	sync.RWMutex
	whichType   int
	path        string
	mtime       time.Time
	size        int64
	url         string
	contentHash uint64
	tag         string
	jwtSecret   string
	bootstrap   []string
	rdata       TextData
}

func NewFromArgs(datatype string, tag string, from string) *DataTable {
	tag = strings.ToUpper(tag)
	if validutil.IsURL(from) {
		return NewDataTable(datatype, tag, WhichTypeUrl, "", from, "")
	}
	if common.FileExists(from) {
		return NewDataTable(datatype, tag, WhichTypePath, from, "", "")
	}
	// log.Println("[Warn]", datatype, tag, from, "no data loaded")
	return NewDataTable(datatype, tag, WhichTypeInline, "", "", from)
}

func NewDataTable(datatype string, tag string, wtype int, path, url string, inline string) *DataTable {
	dt := &DataTable{
		RWMutex:     sync.RWMutex{},
		tag:         tag,
		whichType:   wtype,
		path:        path,
		mtime:       time.Time{},
		size:        0,
		url:         url,
		contentHash: 0,
	}
	switch datatype {
	case DateTypeKeywordTable:
		dt.rdata = newKeywordData(tag)
	case DateTypeNetlistTable:
		dt.rdata = newNetlistData(tag)
	case DateTypeDomainlistTable:
		dt.rdata = newDomainData(tag)
	case DateTypeEcsTable:
		dt.rdata = NewEcsData(tag)
	default:
		return nil
	}
	if inline != "" && strings.HasPrefix(inline, tag) {
		dt.rdata.ParseLines(strings.Fields(inline), false)
	}
	return dt
}

func (kt *DataTable) Reset() {
	kt.rdata.Reset()
}

func (kt *DataTable) Match(name string) bool {
	return kt.rdata.Match(name)
}

func (kt *DataTable) Len() int {
	return kt.rdata.Len()
}

func (kt *DataTable) LoadFromFile() {
	if kt.whichType != WhichTypePath || len(kt.path) == 0 {
		return
	}
	file, err := os.Open(kt.path)
	if err != nil {
		fmt.Printf("DataTable file read error %s", kt.path)
		return
	}

	stat, err := file.Stat()
	if err == nil {
		kt.RLock()
		mtime := kt.mtime
		size := kt.size
		kt.RUnlock()

		if stat.ModTime() == mtime && stat.Size() == size {
			return
		}
	} else {
		// Proceed parsing anyway
		fmt.Printf("%v", err)
	}

	err = kt.rdata.ParseFile(file)
	if err != nil {
		fmt.Printf("parse rdata err %s", err.Error())
		return
	}
	kt.Lock()
	kt.mtime = stat.ModTime()
	kt.size = stat.Size()
	kt.Unlock()
}

func (kt *DataTable) LoadFromUrl() {
	if kt.whichType != WhichTypeUrl || len(kt.url) == 0 {
		return
	}

	token, _ := common.CreateToken(kt.jwtSecret)
	if token != "" {
		if strings.Contains(kt.url, "?") {
			kt.url = kt.url + "&token=" + token
		} else {
			kt.url = kt.url + "?token=" + token
		}
	}

	content, err := httpc.Get(kt.url, nil, kt.bootstrap, time.Second*30)
	if err != nil {
		fmt.Printf("Failed to update %q, err: %v", kt.url, err)
		return
	}
	contentStr := string(content)

	kt.RLock()
	contentHash := kt.contentHash
	kt.RUnlock()
	contentHash1 := common.StringHash(contentStr)
	if contentHash1 == contentHash {
		return
	}
	lines := strings.Split(contentStr, "\n")
	kt.rdata.ParseLines(lines, true)
	kt.Lock()
	kt.contentHash = contentHash1
	kt.Unlock()
}

func (kt *DataTable) LoadFromInline(ws []string) {
	if kt.whichType != WhichTypeInline || len(ws) == 0 {
		return
	}
	kt.Lock()
	defer kt.Unlock()
	kt.rdata.ParseInline(ws)
}

func (kt *DataTable) LoadAll() {
	switch kt.whichType {
	case WhichTypePath:
		kt.LoadFromFile()
	case WhichTypeUrl:
		kt.LoadFromUrl()
	}
}

func (kt *DataTable) String() string {
	sb := strings.Builder{}
	sb.WriteString("DataTable:{")
	sb.WriteString(kt.rdata.LessString())
	sb.WriteString("}")
	return sb.String()
}

func (kt *DataTable) GetData() TextData {
	return kt.rdata
}

func (kt *DataTable) SetJwtSecret(s string) {
	kt.jwtSecret = s
}

func (kt *DataTable) SetBootstrap(bs []string) {
	kt.bootstrap = bs
}
