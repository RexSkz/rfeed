package feed

import (
	"crypto/md5"
	"fmt"
	"testing"

	strip "github.com/grokify/html-strip-tags-go"
	"github.com/mmcdole/gofeed"
	conf "github.com/vasilishin/rfeed/config"
)

var rssXML = `
<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">

<channel>
	<title>Title</title>
	<link>https://example.com/</link>
	<description>Description</description>
	<generator>https://example.com/</generator>
	<image>
		<title>Image Title</title>
		<link>https://example.com/</link>
		<url>https://example.com/icon.png</url>
	</image>
	<item>
		<title>Item Title 1</title>
		<link>https://example1.com/</link>
		<description>Item description 1</description>
		<category><![CDATA[Test1]]></category>
		<category><![CDATA[Test2]]></category>
	</item>
	<item>
		<title>Item Title 2</title>
		<link>https://example2.com/</link>
		<description>Item description 2</description>
		<category><![CDATA[Test1]]></category>
		<category><![CDATA[Test3]]></category>
	</item>
</channel>

</rss>`

func init() {
	var err error
	// Read settings from config file
	conf.Settings, err = conf.NewSettings("config", "..")
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}
}

func getFeed(data string) (*gofeed.Feed, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseString(data)
	return feed, err
}

func TestNewItem(t *testing.T) {
	feed, _ := getFeed(rssXML)
	for _, feedItem := range feed.Items {
		item := NewItem(feed, feedItem)
		authorTitle, authorImg := getImage(feed.Image)
		_, itemImg := getImage(feedItem.Image)
		// Chack Item Author Title value
		if item.Author.Title != authorTitle {
			t.Errorf(
				"Item.Author.Title was incorrect, got: %v, want: %v",
				item.Author.Title,
				authorTitle,
			)
			// Chack Item Author Image value
		} else if item.Author.Image != authorImg {
			t.Errorf(
				"Item.Author.Image was incorrect, got: %v, want: %v",
				item.Author.Image, authorImg,
			)
			// Chack Item Author Link value
		} else if item.Author.Link != feed.Link {
			t.Errorf(
				"Item.Author.Link was incorrect, got: %v, want: %v",
				item.Author.Link, feed.Link,
			)
			// Chack Item Title value without html tags
		} else if item.Title != strip.StripTags(feedItem.Title) {
			t.Errorf(
				"Item.Title was incorrect, got: %v, want: %v",
				item.Title,
				strip.StripTags(feedItem.Title),
			)
			// Chack Item Description value without html tags
		} else if item.Description != strip.StripTags(feedItem.Description) {
			t.Errorf(
				"Item.Description was incorrect, got: %v, want: %v",
				item.Description,
				strip.StripTags(feedItem.Description),
			)
			// Chack Item Link value
		} else if item.Link != feedItem.Link {
			t.Errorf(
				"Item.Link was incorrect, got: %v, want: %v",
				item.Link, feedItem.Link,
			)
			// Chack Item Image value
		} else if item.Image != itemImg {
			t.Errorf(
				"Item.Image was incorrect, got: %v, want: %v",
				item.Image, itemImg,
			)
		}
	}
}

func TestGetMD5Hash(t *testing.T) {
	// return md5 hash from string
	MD5Hash := func(s string) []byte {
		hasher := md5.New()
		hasher.Write([]byte(s))
		return hasher.Sum(nil)
	}

	feed, _ := getFeed(rssXML)
	for _, feedItem := range feed.Items {
		item := NewItem(feed, feedItem)
		itemHash := item.GetMD5Hash()

		hash := MD5Hash(item.Link)
		if string(itemHash) != string(hash) {
			t.Errorf("Item md5 hash was incorrect, got: %v, want: %v", itemHash, hash)
		}

		hash = MD5Hash(item.Link + "test")
		if string(itemHash) == string(hash) {
			t.Errorf("Item md5 hash was incorrect, got: %v, want: %v", itemHash, hash)
		}
	}
}

func TestSkipItem(t *testing.T) {
	// restore settings tags after this test
	saved := conf.Settings.Tags
	defer func() {
		conf.Settings.Tags = saved
	}()

	feed, _ := getFeed(rssXML)
	tables := []struct {
		Item   *gofeed.Item
		Tags   []string
		Wanted bool
	}{
		{feed.Items[0], []string{"Test1", "Test2"}, false},
		{feed.Items[0], []string{"Test1", "Test3"}, false},
		{feed.Items[0], []string{"Test4", "Test3"}, true},
		{feed.Items[1], []string{"Test1", "Test3"}, false},
		{feed.Items[1], []string{"Test1", "Test2"}, false},
		{feed.Items[1], []string{"Test4", "Test2"}, true},
	}

	for _, table := range tables {
		conf.Settings.Tags = table.Tags
		if res := SkipItem(table.Item); res != table.Wanted {
			t.Errorf(
				"Skip result incorrect, got: %v, want: %v, tags: %v, item: %v",
				res, table.Wanted, table.Tags, table.Item.Title,
			)
		}
	}
}

func TestFindItems(t *testing.T) {
	// restore settings tags after this test
	saved := conf.Settings.Tags
	defer func() {
		conf.Settings.Tags = saved
	}()

	tables := []struct {
		Tags   []string
		Wanted int
	}{
		{[]string{"Test1"}, 2},
		{[]string{"Test2"}, 1},
		{[]string{"Test1", "Test2"}, 2},
		{[]string{"Test3"}, 1},
		{[]string{"Test1", "Test3"}, 2},
		{[]string{"Test4"}, 0},
	}

	feed, _ := getFeed(rssXML)
	for _, table := range tables {
		conf.Settings.Tags = table.Tags
		items := FindItems(feed)
		// Chack found items length
		if len(items) != table.Wanted {
			t.Errorf(
				"Items found len was incorrect, got: %v, want: %v, tags: %v",
				len(items), table.Wanted, table.Tags,
			)
		}
	}
}
