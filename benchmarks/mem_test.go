/*
 * Copyright (c) 2018-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package benchmarks

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/linkedin/goavro"
	"github.com/stretchr/testify/require"
	"github.com/untillpro/dynobuffers"
)

func Test_MemFewArticleFields_Avro(t *testing.T) {
	require := require.New(t)
	schemaStr, err := os.ReadFile("article-nullable.avsc")
	require.NoError(err)

	codec, err := goavro.NewCodec(string(schemaStr))
	require.NoError(err)

	data := make(map[string]interface{})
	data["quantity"] = goavro.Union("int", int(123))
	data["purchase_price"] = goavro.Union("float", float32(0.123))
	data["id"] = goavro.Union("long", int64(123))

	buf := make([]byte, 0)
	bytes, err := codec.BinaryFromNative(buf, data)
	require.NoError(err)

	log.Println("Buffer len", len(bytes))
}

func Test_MemFewArticleFields_Dyno(t *testing.T) {
	require := require.New(t)
	s := getArticleSchemeDynoBuffer()
	bf := dynobuffers.NewBuffer(s)
	bf.Set("quantity", 10)
	bf.Set("purchase_price", 12.34)
	bytes, err := bf.ToBytes()
	require.NoError(err)
	log.Println("Buffer len for quantity, purchase_price:", len(bytes))

	bf.Set("id", 5678)
	bytes, err = bf.ToBytes()
	require.NoError(err)
	log.Println("Buffer len for quantity, purchase_price, id:", len(bytes))

	bf.Set("article_number", 9)
	bytes, err = bf.ToBytes()
	require.NoError(err)
	log.Println("Buffer len for quantity, purchase_price, id, article_number:", len(bytes))

	bf.Release()
	require.Zero(dynobuffers.GetObjectsInUse())
}

func Test_MemAllArticleFields_Avro(t *testing.T) {
	require := require.New(t)
	schemaStr, err := ioutil.ReadFile("article.avsc")
	require.NoError(err)

	codec, err := goavro.NewCodec(string(schemaStr))
	require.NoError(err)

	articleData, err := ioutil.ReadFile("articleData.json")
	require.NoError(err)

	native, _, err := codec.NativeFromTextual(articleData)
	require.NoError(err)

	bytes, err := codec.BinaryFromNative(nil, native)
	require.NoError(err)

	//	ioutil.WriteFile("article.avro", bytes, 0644)

	log.Println("MemAllArticleFields_Avro:", len(bytes))
}

func Test_MemAllArticleFields_Dyno(t *testing.T) {
	require := require.New(t)
	s := getArticleSchemeDynoBuffer()
	bf := dynobuffers.NewBuffer(s)
	fillArticleDynoBuffer(bf)
	bytes, err := bf.ToBytes()
	require.NoError(err)

	log.Println("MemAllArticleFields_Dyno:", len(bytes))
	bf.Release()
	require.Zero(dynobuffers.GetObjectsInUse())
}
