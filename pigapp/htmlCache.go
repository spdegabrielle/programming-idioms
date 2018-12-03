package main

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"strconv"
	"time"

	. "github.com/Deleplace/programming-idioms/pig"

	"golang.org/x/net/context"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/taskqueue"
)

// Sometimes we want to saved whole blocks of template-generated HTML
// into HTML, and serve it again later.

// htmlCacheRead returns previously saved bytes for this key,
// It returns nil if not found, or expired, or on cache error.
//
// There is no guarantee that previously cached data will be found,
// because cache entries may vanish anytime, even before expiration.
func htmlCacheRead(c context.Context, key string) []byte {
	value, err := cache.read(c, key)
	if err != nil {
		errorf(c, "Reading cache for %q: %v", key, err)
	}
	return value
}

// htmlCacheWrite saves bytes for given key.
// Failures are ignored.
func htmlCacheWrite(c context.Context, key string, data []byte, duration time.Duration) {
	err := cache.write(c, key, data, duration)
	if err != nil {
		errorf(c, "Writing cache for %q: %v", key, err)
	}
}

// Data changes should lead to cache entries invalidation.
func htmlCacheEvict(c context.Context, key string) {
	err := cache.evict(c, key)
	if err != nil {
		errorf(c, "Evicting cache for %q: %v", key, err)
	}
	// See also htmlUncacheIdiomAndImpls
}

// When expected data may be >1MB.
func htmlCacheZipRead(c context.Context, key string) []byte {
	zipdata := htmlCacheRead(c, key)
	if zipdata == nil {
		return nil
	}
	zipbuffer := bytes.NewBuffer(zipdata)
	zipreader, err := gzip.NewReader(zipbuffer)
	if err != nil {
		errorf(c, "Reading zip cached entry %q: %v", key, err)
		// Ignore failure
		return nil
	}
	buffer, err := ioutil.ReadAll(zipreader)
	if err != nil {
		errorf(c, "Reading zip cached entry %q: %v", key, err)
	}
	debugf(c, "Reading %d bytes out of %d gzip bytes for entry %q", len(buffer), len(zipdata), key)
	return buffer
}

// When expected data may be >1MB.
func htmlCacheZipWrite(c context.Context, key string, data []byte, duration time.Duration) {
	var zipbuffer bytes.Buffer
	zipwriter := gzip.NewWriter(&zipbuffer)
	_, err := zipwriter.Write(data)
	if err != nil {
		errorf(c, "Writing zip cached entry %q: %v", key, err)
		// Ignore failure
		return
	}
	_ = zipwriter.Close()
	debugf(c, "Writing %d gzip bytes out of %d data bytes for entry %q", zipbuffer.Len(), len(data), key)
	htmlCacheWrite(c, key, zipbuffer.Bytes(), duration)
}

func htmlUncacheIdiomAndImpls(c context.Context, idiom *Idiom) {
	//
	// There are only two hard things in Computer Science: cache invalidation and naming things.
	//
	infof(c, "Evicting HTML cached pages for idiom %d %q", idiom.Id, idiom.Title)

	cachekeys := make([]string, 0, 1+len(idiom.Implementations))
	cachekeys = append(cachekeys, NiceIdiomRelativeURL(idiom))
	for _, impl := range idiom.Implementations {
		cachekeys = append(cachekeys, NiceImplRelativeURL(idiom, impl.Id, impl.LanguageName))
	}
	for _, key := range cachekeys {
		err := cache.evict(c, key)
		// Note that here we're ignoring potential cache errors
		_ = err
	}
}

func htmlRecacheNowAndTomorrow(c context.Context, idiomID int) error {
	debugf(c, "Creating html recache tasks for idiom %d", idiomID)
	// These 2 task submissions may take several 10s of ms,
	// thus we decide to submit them as a small batch.

	// Now
	t1, err1 := recacheHtmlIdiom.Task(idiomID)
	if err1 != nil {
		return err1
	}

	// Tomorrow
	t2, err2 := recacheHtmlIdiom.Task(idiomID)
	if err2 != nil {
		return err2
	}
	t2.Delay = 24*time.Hour + 10*time.Minute

	_, err := taskqueue.AddMulti(c, []*taskqueue.Task{t1, t2}, "")
	return err
}

var recacheHtmlIdiom, recacheHtmlImpl *delay.Function

func init() {
	recacheHtmlIdiom = delay.Func("recache-html-idiom", func(c context.Context, idiomID int) {
		infof(c, "Start recaching HTML for idiom %d", idiomID)
		_, idiom, err := dao.getIdiom(c, idiomID)
		if err != nil {
			errorf(c, "recacheHtmlIdiom: %v", err)
			return
		}

		path := NiceIdiomRelativeURL(idiom)
		var buffer bytes.Buffer
		vars := map[string]string{
			"idiomId":    strconv.Itoa(idiomID),
			"idiomTitle": uriNormalize(idiom.Title),
		}
		err = generateIdiomDetailPage(c, &buffer, vars)
		if err != nil {
			errorf(c, "recacheHtmlIdiom: %v", err)
			return
		}
		htmlCacheWrite(c, path, buffer.Bytes(), 24*time.Hour)

		// Then, create async task for each impl to be HTML-recached
		for _, impl := range idiom.Implementations {
			implPath := NiceImplRelativeURL(idiom, impl.Id, impl.LanguageName)
			recacheHtmlImpl.Call(c, implPath, idiom.Id, idiom.Title, impl.Id, impl.LanguageName)
		}

	})

	recacheHtmlImpl = delay.Func("recache-html-impl", func(
		c context.Context,
		implPath string,
		idiomID int,
		idiomTitle string,
		implID int,
		implLang string,
	) {
		infof(c, "Recaching HTML for %s", implPath)
		// TODO call idiomDetail(fakeWriter, fakeRequest)

		var buffer bytes.Buffer
		vars := map[string]string{
			"idiomId":    strconv.Itoa(idiomID),
			"idiomTitle": uriNormalize(idiomTitle),
			"implId":     strconv.Itoa(implID),
			"implLang":   implLang,
		}
		err := generateIdiomDetailPage(c, &buffer, vars)
		if err != nil {
			errorf(c, "recacheHtmlImpl: %v", err)
			return
		}
		htmlCacheWrite(c, implPath, buffer.Bytes(), 24*time.Hour)
	})
}