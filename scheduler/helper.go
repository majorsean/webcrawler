/*
* @Author: wangshuo
* @Date:   2017-04-12 12:31:53
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-19 11:22:29
 */

package scheduler

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	anlz "webcrawler/analyzer"
	base "webcrawler/base"
	dl "webcrawler/downloader"
	ipl "webcrawler/itempipeline"
	mdw "webcrawler/middleware"
)

func generateChannelManager(channelArgs base.ChannelArgs) mdw.ChannelManager {
	return mdw.NewChannelManager(channelArgs)
}

func generateAnalyzerPool(total uint32) (anlz.AnalyzerPool, error) {
	analyerPool, err := anlz.NewAnalyzerPool(
		total,
		func() anlz.Analyzer {
			return anlz.NewAnalyzer()
		},
	)
	if err != nil {
		return nil, err
	}
	return analyerPool, nil
}

func generatePageDownloaderPool(total uint32, httpClientGenerator GenhttpClient) (dl.PageDownloaderPool, error) {
	dlPool, err := dl.NewPageDownloaderPool(
		total,
		func() dl.PageDownloader {
			return dl.NewPageDownloader(httpClientGenerator())
		},
	)
	if err != nil {
		return nil, err
	}
	return dlPool, nil
}

func generateItemProcessors(itemProcessors []ipl.ProcessItem) ipl.ItemPipeline {
	return ipl.NewItemPipeline(itemProcessors)
}

var regexpForIp = regexp.MustCompile(`((?:(?:25[0-5]|2[0-4]\d|[01]?\d?\d)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d?\d))`)

var regexpForDomain = []*regexp.Regexp{
	regexp.MustCompile(`\.(com|com\.\w{2})$`),
	regexp.MustCompile(`\.(gov|gov\.\w{2})$`),
	regexp.MustCompile(`\.(net|net\.\w{2})$`),
	regexp.MustCompile(`\.(org|org\.\w{2})$`),
	// *.xx
	regexp.MustCompile(`\.me$`),
	regexp.MustCompile(`\.biz$`),
	regexp.MustCompile(`\.info$`),
	regexp.MustCompile(`\.name$`),
	regexp.MustCompile(`\.mobi$`),
	regexp.MustCompile(`\.so$`),
	regexp.MustCompile(`\.asia$`),
	regexp.MustCompile(`\.tel$`),
	regexp.MustCompile(`\.tv$`),
	regexp.MustCompile(`\.cc$`),
	regexp.MustCompile(`\.co$`),
	regexp.MustCompile(`\.\w{2}$`),
}

func generateCode(prefix string, id uint32) string {
	return fmt.Sprintf("%s-%d", prefix, id)
}

func parseCode(code string) []string {
	result := make([]string, 2)
	var codePrefix string
	var id string
	index := strings.Index(code, "-")
	if index > 0 {
		codePrefix = code[:index]
		id = code[index+1:]
	} else {
		codePrefix = code
	}
	result[0] = codePrefix
	result[1] = id
	return result
}

func getPrimaryDomain(host string) (string, error) {
	fmt.Println(host)
	host = strings.TrimSpace(host)
	if host == "" {
		return "", errors.New("The host is empty!")
	}
	if regexpForIp.MatchString(host) {
		return host, nil
	}
	var suffixMatch int
	for _, reg := range regexpForDomain {
		res := reg.FindStringIndex(host)
		if len(res) > 0 {
			suffixMatch = res[0]
			break
		}
	}
	if suffixMatch > 0 {
		index := strings.LastIndex(host[:suffixMatch], ".")
		if index > 0 {
			index++
		} else {
			index = 0
		}
		return host[index:], nil
	} else {
		return "", errors.New("Unrecognized host!")
	}
}
