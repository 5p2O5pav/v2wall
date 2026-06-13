package ipgeo

import (
	"net"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

type Searcher struct {
	v4Searcher *xdb.Searcher
	v6Searcher *xdb.Searcher
	enableV4   bool
	enableV6   bool
}

func NewSearcher(v4Path, v6Path string, enableV4, enableV6 bool) (*Searcher, error) {
	s := &Searcher{
		enableV4: enableV4,
		enableV6: enableV6,
	}

	if enableV4 {
		content, err := xdb.LoadContentFromFile(v4Path)
		if err != nil {
			return nil, err
		}
		searcher, err := xdb.NewWithBuffer(xdb.IPv4, content)
		if err != nil {
			return nil, err
		}
		s.v4Searcher = searcher
	}

	if enableV6 {
		content, err := xdb.LoadContentFromFile(v6Path)
		if err != nil {
			return nil, err
		}
		searcher, err := xdb.NewWithBuffer(xdb.IPv6, content)
		if err != nil {
			return nil, err
		}
		s.v6Searcher = searcher
	}
	return s, nil
}

func (s *Searcher) Search(ipStr string) (string, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", net.InvalidAddrError("invalid IP")
	}
	if ip.To4() != nil {
		if !s.enableV4 || s.v4Searcher == nil {
			return "", nil // IPv4 查询未启用，返回空
		}
		return s.v4Searcher.Search(ipStr)
	} else {
		if !s.enableV6 || s.v6Searcher == nil {
			return "", nil
		}
		return s.v6Searcher.Search(ipStr)
	}
}

func (s *Searcher) Close() {
	if s.v4Searcher != nil {
		s.v4Searcher.Close()
	}
	if s.v6Searcher != nil {
		s.v6Searcher.Close()
	}
}
