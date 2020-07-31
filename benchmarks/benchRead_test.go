/*
 * Copyright (c) 2019-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package benchmarks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/Yohanson555/dynobuffers"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/linkedin/goavro"
	"github.com/stretchr/testify/require"
)

func BenchmarkReadDynoBuffersSimpleTypedReadString(b *testing.B) {
	s := getSimpleScheme()
	bf := dynobuffers.NewBuffer(s)
	bf.Set("name", "cola")
	bf.Set("price", float32(0.123))
	bf.Set("quantity", int32(42))
	bytes, _ := bf.ToBytes()

	b.ResetTimer()
	sum := float32(0)
	for i := 0; i < b.N; i++ {
		bf := dynobuffers.ReadBuffer(bytes, s)

		price, _ := bf.GetFloat("price")
		quantity, _ := bf.GetInt("quantity")
		_, _ = bf.GetString("name")
		sum += price * float32(quantity)

		bf.Release()
	}
}

func BenchmarkReadDynoBuffersSimpleUntyped(b *testing.B) {
	s := getSimpleScheme()
	bf := dynobuffers.NewBuffer(s)
	bf.Set("name", "cola")
	bf.Set("price", float32(0.123))
	bf.Set("quantity", int32(42))
	bytes, _ := bf.ToBytes()

	b.ResetTimer()
	sum := float32(0)
	for i := 0; i < b.N; i++ {
		bf := dynobuffers.ReadBuffer(bytes, s)
		price := bf.Get("price").(float32)
		quantity := bf.Get("quantity").(int32)
		sum += price * float32(quantity)
		bf.Release()
	}
}

func BenchmarkReadDynoBuffersApplyJSONArraysAllTypesAppendNoNested(t *testing.B) {
	arraysAllTypesYaml := `
ints..: int
longs..: long
floats..: float
doubles..: double
strings..: string
boolTrues..: bool
boolFalses..: bool
bytes..: byte
intsObj..:
  int: int
`
	s, err := dynobuffers.YamlToScheme(arraysAllTypesYaml)
	if err != nil {
		t.Fatal(err)
	}
	b := dynobuffers.NewBuffer(s)
	bytes, err := b.ApplyJSONAndToBytes([]byte(`{"ints":[1,2],"longs":[3,4],"floats":[0.123,0.124],"doubles":[0.125,0.126],"strings":["str1","str2"],"boolTrues":[true,false],"boolFalses":[false,true],"bytes":"BQY="}`))
	if err != nil {
		t.Fatal(err)
	}

	dest := map[string]interface{}{}
	jsonBytes := []byte(`{"ints":[-1,-2],"longs":[-3,-4],"floats":[-0.123,-0.124],"doubles":[-0.125,-0.126],"strings":["","str4"],"boolTrues":[true,true],"boolFalses":[false,false],"bytes":"BQY="}`)
	err = json.Unmarshal(jsonBytes, &dest)
	if err != nil {
		t.Fatal(err)
	}

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		bf := dynobuffers.ReadBuffer(bytes, s)
		err = bf.ApplyMap(dest)
		if err != nil {
			t.Fatal(err)
		}
		_, err = bf.ToBytes()
		if err != nil {
			t.Fatal(err)
		}

		bf.Release()
	}
}

func BenchmarkReadDynoBuffersApplyJSONArraysAllTypesAppendNested(t *testing.B) {
	arraysAllTypesYaml := `
ints..: int
longs..: long
floats..: float
doubles..: double
strings..: string
boolTrues..: bool
boolFalses..: bool
bytes..: byte
intsObj..:
  int: int
`
	s, err := dynobuffers.YamlToScheme(arraysAllTypesYaml)
	if err != nil {
		t.Fatal(err)
	}
	b := dynobuffers.NewBuffer(s)
	bytes, err := b.ApplyJSONAndToBytes([]byte(`{"ints":[1,2],"longs":[3,4],"floats":[0.123,0.124],"doubles":[0.125,0.126],"strings":["str1","str2"],"boolTrues":[true,false],"boolFalses":[false,true],"bytes":"BQY=","intsObj":[{"int":7},{"int":8}]}`))
	if err != nil {
		t.Fatal(err)
	}

	dest := map[string]interface{}{}
	jsonBytes := []byte(`{"ints":[-1,-2],"longs":[-3,-4],"floats":[-0.123,-0.124],"doubles":[-0.125,-0.126],"strings":["","str4"],"boolTrues":[true,true],"boolFalses":[false,false],"bytes":"BQY=","intsObj":[{"int":-7},{"int":-8}]}`)
	err = json.Unmarshal(jsonBytes, &dest)
	if err != nil {
		t.Fatal(err)
	}

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		bf := dynobuffers.ReadBuffer(bytes, s)
		err = bf.ApplyMap(dest)
		if err != nil {
			t.Fatal(err)
		}
		_, err = bf.ToBytes()
		if err != nil {
			t.Fatal(err)
		}
		bf.Release()
	}
}

func Benchmark_ReadSimple_Avro(b *testing.B) {
	codec, err := goavro.NewCodec(`
		{"namespace":"unTill",
		"type": "record",
		"name": "OrderItem",
		"fields": [
			{"name": "name", "type": "string"},
			{"name": "price", "type": ["float", "null"]},
			{"name": "quantity", "type": "int", "default": 0}
		]}
	`)
	if err != nil {
		b.Fatal(err)
	}
	data := map[string]interface{}{
		"name":     string("cola"),
		"price":    goavro.Union("float", float32(0.123)),
		"quantity": 1,
	}
	bytes, err := codec.BinaryFromNative(nil, data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	var sum float64
	for i := 0; i < b.N; i++ {
		native, _, err := codec.NativeFromBinary(bytes)
		if err != nil {
			b.Fatal(err)
		}
		decoded := native.(map[string]interface{})
		price := float64(decoded["price"].(map[string]interface{})["float"].(float32))
		sum += price * float64(decoded["quantity"].(int32))
	}
}
func Benchmark_ReadSimple_Dyno_Typed(b *testing.B) {
	s := getSimpleScheme()
	bf := dynobuffers.NewBuffer(s)
	bf.Set("name", "cola")
	bf.Set("price", float32(0.123))
	bf.Set("quantity", int32(42))
	bytes, _ := bf.ToBytes()

	b.ResetTimer()
	sum := float32(0)
	for i := 0; i < b.N; i++ {
		bf := dynobuffers.ReadBuffer(bytes, s)
		price, _ := bf.GetFloat("price")
		quantity, _ := bf.GetInt("quantity")
		sum += price * float32(quantity)
		bf.Release()
	}
}
func Benchmark_ReadSimple_Flat(b *testing.B) {
	bl := flatbuffers.NewBuilder(0)
	colaName := bl.CreateString("cola")
	SaleStart(bl)
	SaleAddName(bl, colaName)
	SaleAddPrice(bl, 0.123)
	SaleAddQuantity(bl, int32(1))
	sale := SaleEnd(bl)
	bl.Finish(sale)
	bytes := bl.FinishedBytes()

	b.ResetTimer()
	var sum float64
	for i := 0; i < b.N; i++ {
		saleNew := GetRootAsSale(bytes, 0)
		sum += float64(saleNew.Price()) * float64(saleNew.Quantity())
	}
}
func Benchmark_ReadSimple_Flat_String(b *testing.B) {
	bl := flatbuffers.NewBuilder(0)
	colaName := bl.CreateString("cola")
	SaleStart(bl)
	SaleAddName(bl, colaName)
	SaleAddPrice(bl, 0.123)
	SaleAddQuantity(bl, int32(1))
	sale := SaleEnd(bl)
	bl.Finish(sale)
	bytes := bl.FinishedBytes()

	b.ResetTimer()
	var sum float64
	for i := 0; i < b.N; i++ {
		saleNew := GetRootAsSale(bytes, 0)
		sum += float64(saleNew.Price()) * float64(saleNew.Quantity())
		_ = string(saleNew.Name())
	}
}

func Benchmark_ReadSimple_Json(b *testing.B) {
	var data map[string]interface{}
	data = map[string]interface{}{
		"name":     "cola",
		"price":    0.123,
		"quantity": int32(1),
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	b.ResetTimer()

	var sum float64
	for i := 0; i < b.N; i++ {
		data = map[string]interface{}{}
		err = json.Unmarshal(bytes, &data)
		if err != nil {
			fmt.Println(err)
		}
		price := float64(data["price"].(float64))
		sum += price * data["quantity"].(float64)
	}
}

func Benchmark_ReadFewArticleFields_Avro(b *testing.B) {
	schemaStr, err := ioutil.ReadFile("article.avsc")
	require.Nil(b, err)

	codec, err := goavro.NewCodec(string(schemaStr))
	require.Nil(b, err)

	articleData, err := ioutil.ReadFile("articleData.json")
	require.Nil(b, err)

	native, _, err := codec.NativeFromTextual(articleData)
	require.Nil(b, err)

	bytes, err := codec.BinaryFromNative(nil, native)
	require.Nil(b, err)

	b.ResetTimer()
	sum := float64(0)
	for i := 0; i < b.N; i++ {
		native, _, err = codec.NativeFromBinary(bytes)
		if err != nil {
			b.Fatal(err)
		}
		decoded := native.(map[string]interface{})
		price := float64(decoded["purchase_price"].(float32))
		sum += price * float64(decoded["quantity"].(int32))
	}

}
func Benchmark_ReadFewArticleFields_Dyno_Typed(b *testing.B) {
	s := getArticleSchemeDynoBuffer()
	bf := dynobuffers.NewBuffer(s)
	fillArticleDynoBuffer(bf)
	bytes, _ := bf.ToBytes()
	b.ResetTimer()
	sum := float64(0)
	for i := 0; i < b.N; i++ {
		bf := dynobuffers.ReadBuffer(bytes, s)
		q, _ := bf.GetInt("quantity")
		price, _ := bf.GetFloat("purchase_price")
		sum += float64(float32(q) * price)
		bf.Release()
	}
}
func Benchmark_ReadFewArticleFields_Flat(b *testing.B) {
	bl := flatbuffers.NewBuilder(0)
	a := fillArticleFlatBuffers(bl)
	bl.Finish(a)
	bytes := bl.FinishedBytes()

	b.ResetTimer()
	sum := float64(0)
	for i := 0; i < b.N; i++ {
		a := GetRootAsArticle(bytes, 0)
		sum += float64(a.PurchasePrice()) * float64(a.Quantity())
	}
}

func Benchmark_ReadFewArticleFields_Json(b *testing.B) {
	s := getArticleSchemeDynoBuffer()
	bf := dynobuffers.NewBuffer(s)
	fillArticleDynoBuffer(bf)
	jsonStr := bf.ToJSON()
	dest := map[string]interface{}{}
	b.ResetTimer()
	sum := float64(0)
	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(jsonStr), &dest)
		_ = dest["id_courses"]
		q := dest["quantity"].(float64)
		price := dest["purchase_price"].(float64)
		sum += q * price
	}
}

func Benchmark_ReadAllArticleFields_Avro(b *testing.B) {
	schemaStr, err := ioutil.ReadFile("article.avsc")
	require.Nil(b, err)

	codec, err := goavro.NewCodec(string(schemaStr))
	require.Nil(b, err)

	articleData, err := ioutil.ReadFile("articleData.json")
	require.Nil(b, err)

	native, _, err := codec.NativeFromTextual(articleData)
	require.Nil(b, err)

	bytes, err := codec.BinaryFromNative(nil, native)
	require.Nil(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		native, _, err = codec.NativeFromBinary(bytes)
		if err != nil {
			b.Fatal(err)
		}
		decoded := native.(map[string]interface{})
		_ = decoded["id"]
		_ = decoded["article_number"]
		_ = decoded["name"]
		_ = decoded["internal_name"]
		_ = decoded["article_manual"]
		_ = decoded["article_hash"]
		_ = decoded["id_courses"]
		_ = decoded["id_departament"]
		_ = decoded["pc_bitmap"]
		_ = decoded["pc_color"]
		_ = decoded["pc_text"]
		_ = decoded["pc_font_name"]
		_ = decoded["pc_font_size"]
		_ = decoded["pc_font_attr"]
		_ = decoded["pc_font_color"]
		_ = decoded["rm_text"]
		_ = decoded["rm_font_size"]
		_ = decoded["id_packing"]
		_ = decoded["id_commission"]
		_ = decoded["id_promotions"]
		_ = decoded["savepoints"]
		_ = decoded["quantity"]
		_ = decoded["hideonhold"]
		_ = decoded["barcode"]
		_ = decoded["time_active"]
		_ = decoded["aftermin"]
		_ = decoded["periodmin"]
		_ = decoded["roundmin"]
		_ = decoded["id_currency"]
		_ = decoded["control_active"]
		_ = decoded["control_time"]
		_ = decoded["plu_number_vanduijnen"]
		_ = decoded["sequence"]
		_ = decoded["rm_sequence"]
		_ = decoded["purchase_price"]
		_ = decoded["id_vd_group"]
		_ = decoded["menu"]
		_ = decoded["sensitive"]
		_ = decoded["sensitive_option"]
		_ = decoded["daily_stock"]
		_ = decoded["info"]
		_ = decoded["warning_level"]
		_ = decoded["free_after_pay"]
		_ = decoded["id_food_group"]
		_ = decoded["article_type"]
		_ = decoded["id_inventory_item"]
		_ = decoded["id_recipe"]
		_ = decoded["id_unity_sales"]
		_ = decoded["can_savepoints"]
		_ = decoded["show_in_kitchen_screen"]
		_ = decoded["decrease_savepoints"]
		_ = decoded["hht_color"]
		_ = decoded["hht_font_name"]
		_ = decoded["hht_font_size"]
		_ = decoded["hht_font_attr"]
		_ = decoded["hht_font_color"]
		_ = decoded["tip"]
		_ = decoded["id_beco_group"]
		_ = decoded["id_beco_location"]
		_ = decoded["bc_standard_dosage"]
		_ = decoded["bc_alternative_dosage"]
		_ = decoded["bc_disablebalance"]
		_ = decoded["bc_use_locations"]
		_ = decoded["time_rate"]
		_ = decoded["id_free_option"]
		_ = decoded["party_article"]
		_ = decoded["id_pua_groups"]
		_ = decoded["promo"]
		_ = decoded["one_hand_limit"]
		_ = decoded["consolidate_quantity"]
		_ = decoded["consolidate_alias_name"]
		_ = decoded["hq_id"]
		_ = decoded["is_active"]
		_ = decoded["is_active_modified"]
		_ = decoded["is_active_modifier"]
		_ = decoded["rent_price_type"]
		_ = decoded["id_rental_group"]
		_ = decoded["condition_check_in_order"]
		_ = decoded["weight_required"]
		_ = decoded["daily_numeric_1"]
		_ = decoded["daily_numeric_2"]
		_ = decoded["prep_min"]
		_ = decoded["id_article_ksp"]
		_ = decoded["warn_min"]
		_ = decoded["empty_article"]
		_ = decoded["bc_debitcredit"]
		_ = decoded["prep_sec"]
		_ = decoded["id_suppliers"]
		_ = decoded["main_price"]
		_ = decoded["oman_text"]
		_ = decoded["id_age_groups"]
		_ = decoded["surcharge"]
		_ = decoded["info_data"]
		_ = decoded["pos_disabled"]
		_ = decoded["ml_name"]
		_ = decoded["ml_ks_name"]
		_ = decoded["alt_articles"]
		_ = decoded["alt_alias"]
		_ = decoded["need_prep"]
		_ = decoded["auto_onhold"]
		_ = decoded["id_ks_wf"]
		_ = decoded["ks_wf_type"]
		_ = decoded["ask_course"]
		_ = decoded["popup_info"]
		_ = decoded["allow_order_items"]
		_ = decoded["must_combined"]
		_ = decoded["block_discount"]
		_ = decoded["has_default_options"]
		_ = decoded["hht_default_setting"]
		_ = decoded["oman_default_setting"]
		_ = decoded["id_rent_periods"]
		_ = decoded["delay_separate_mins"]
		_ = decoded["id_ksc"]
		_ = decoded["ml_pc_text"]
		_ = decoded["ml_rm_text"]
		_ = decoded["ml_oman_text"]
		_ = decoded["pos_article_type"]
		_ = decoded["single_free_option"]
		_ = decoded["ks_single_item"]
		_ = decoded["allergen"]
		_ = decoded["auto_resetcourse"]
		_ = decoded["block_transfer"]
		_ = decoded["id_size_modifier"]
	}

}
func Benchmark_ReadAllArticleFields_Dyno_Untyped(b *testing.B) {
	s := getArticleSchemeDynoBuffer()
	bf := dynobuffers.NewBuffer(s)
	fillArticleDynoBuffer(bf)
	bytes, _ := bf.ToBytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf := dynobuffers.ReadBuffer(bytes, s)
		bf.Get("id")
		bf.Get("article_number")
		bf.Get("name")
		bf.Get("internal_name")
		bf.Get("article_manual")
		bf.Get("article_hash")
		bf.Get("id_courses")
		bf.Get("id_departament")
		bf.Get("pc_bitmap")
		bf.Get("pc_color")
		bf.Get("pc_text")
		bf.Get("pc_font_name")
		bf.Get("pc_font_size")
		bf.Get("pc_font_attr")
		bf.Get("pc_font_color")
		bf.Get("rm_text")
		bf.Get("rm_font_size")
		bf.Get("id_packing")
		bf.Get("id_commission")
		bf.Get("id_promotions")
		bf.Get("savepoints")
		bf.Get("quantity")
		bf.Get("hideonhold")
		bf.Get("barcode")
		bf.Get("time_active")
		bf.Get("aftermin")
		bf.Get("periodmin")
		bf.Get("roundmin")
		bf.Get("id_currency")
		bf.Get("control_active")
		bf.Get("control_time")
		bf.Get("plu_number_vanduijnen")
		bf.Get("sequence")
		bf.Get("rm_sequence")
		bf.Get("purchase_price")
		bf.Get("id_vd_group")
		bf.Get("menu")
		bf.Get("sensitive")
		bf.Get("sensitive_option")
		bf.Get("daily_stock")
		bf.Get("info")
		bf.Get("warning_level")
		bf.Get("free_after_pay")
		bf.Get("id_food_group")
		bf.Get("article_type")
		bf.Get("id_inventory_item")
		bf.Get("id_recipe")
		bf.Get("id_unity_sales")
		bf.Get("can_savepoints")
		bf.Get("show_in_kitchen_screen")
		bf.Get("decrease_savepoints")
		bf.Get("hht_color")
		bf.Get("hht_font_name")
		bf.Get("hht_font_size")
		bf.Get("hht_font_attr")
		bf.Get("hht_font_color")
		bf.Get("tip")
		bf.Get("id_beco_group")
		bf.Get("id_beco_location")
		bf.Get("bc_standard_dosage")
		bf.Get("bc_alternative_dosage")
		bf.Get("bc_disablebalance")
		bf.Get("bc_use_locations")
		bf.Get("time_rate")
		bf.Get("id_free_option")
		bf.Get("party_article")
		bf.Get("id_pua_groups")
		bf.Get("promo")
		bf.Get("one_hand_limit")
		bf.Get("consolidate_quantity")
		bf.Get("consolidate_alias_name")
		bf.Get("hq_id")
		bf.Get("is_active")
		bf.Get("is_active_modified")
		bf.Get("is_active_modifier")
		bf.Get("rent_price_type")
		bf.Get("id_rental_group")
		bf.Get("condition_check_in_order")
		bf.Get("weight_required")
		bf.Get("daily_numeric_1")
		bf.Get("daily_numeric_2")
		bf.Get("prep_min")
		bf.Get("id_article_ksp")
		bf.Get("warn_min")
		bf.Get("empty_article")
		bf.Get("bc_debitcredit")
		bf.Get("prep_sec")
		bf.Get("id_suppliers")
		bf.Get("main_price")
		bf.Get("oman_text")
		bf.Get("id_age_groups")
		bf.Get("surcharge")
		bf.Get("info_data")
		bf.Get("pos_disabled")
		bf.Get("ml_name")
		bf.Get("ml_ks_name")
		bf.Get("alt_articles")
		bf.Get("alt_alias")
		bf.Get("need_prep")
		bf.Get("auto_onhold")
		bf.Get("id_ks_wf")
		bf.Get("ks_wf_type")
		bf.Get("ask_course")
		bf.Get("popup_info")
		bf.Get("allow_order_items")
		bf.Get("must_combined")
		bf.Get("block_discount")
		bf.Get("has_default_options")
		bf.Get("hht_default_setting")
		bf.Get("oman_default_setting")
		bf.Get("id_rent_periods")
		bf.Get("delay_separate_mins")
		bf.Get("id_ksc")
		bf.Get("ml_pc_text")
		bf.Get("ml_rm_text")
		bf.Get("ml_oman_text")
		bf.Get("pos_article_type")
		bf.Get("single_free_option")
		bf.Get("ks_single_item")
		bf.Get("allergen")
		bf.Get("auto_resetcourse")
		bf.Get("block_transfer")
		bf.Get("id_size_modifier")

		bf.Release()
	}
}

func Benchmark_ReadAllArticleFields_Dyno_Typed(b *testing.B) {
	s := getArticleSchemeDynoBuffer()
	bf := dynobuffers.NewBuffer(s)
	fillArticleDynoBuffer(bf)
	bytes, _ := bf.ToBytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf := dynobuffers.ReadBuffer(bytes, s)
		bf.GetLong("id")
		bf.GetInt("article_number")
		bf.GetString("name")
		bf.GetString("internal_name")
		bf.GetInt("article_manual")
		bf.GetInt("article_hash")
		bf.GetLong("id_courses")
		bf.GetLong("id_departament")
		bf.GetInt("pc_bitmap")
		bf.GetInt("pc_color")
		bf.GetInt("pc_text")
		bf.GetString("pc_font_name")
		bf.GetInt("pc_font_size")
		bf.GetInt("pc_font_attr")
		bf.GetInt("pc_font_color")
		bf.GetString("rm_text")
		bf.GetInt("rm_font_size")
		bf.GetInt("id_packing")
		bf.GetLong("id_commission")
		bf.GetLong("id_promotions")
		bf.GetInt("savepoints")
		bf.GetInt("quantity")
		bf.GetInt("hideonhold")
		bf.GetInt("barcode")
		bf.GetInt("time_active")
		bf.GetInt("aftermin")
		bf.GetInt("periodmin")
		bf.GetInt("roundmin")
		bf.GetLong("id_currency")
		bf.GetInt("control_active")
		bf.GetInt("control_time")
		bf.GetInt("plu_number_vanduijnen")
		bf.GetInt("sequence")
		bf.GetInt("rm_sequence")
		bf.GetFloat("purchase_price")
		bf.GetLong("id_vd_group")
		bf.GetInt("menu")
		bf.GetInt("sensitive")
		bf.GetInt("sensitive_option")
		bf.GetInt("daily_stock")
		bf.GetString("info")
		bf.GetInt("warning_level")
		bf.GetInt("free_after_pay")
		bf.GetLong("id_food_group")
		bf.GetInt("article_type")
		bf.GetLong("id_inventory_item")
		bf.GetLong("id_recipe")
		bf.GetLong("id_unity_sales")
		bf.GetInt("can_savepoints")
		bf.GetBool("show_in_kitchen_screen")
		bf.GetInt("decrease_savepoints")
		bf.GetInt("hht_color")
		bf.GetString("hht_font_name")
		bf.GetInt("hht_font_size")
		bf.GetInt("hht_font_attr")
		bf.GetInt("hht_font_color")
		bf.GetInt("tip")
		bf.GetLong("id_beco_group")
		bf.GetLong("id_beco_location")
		bf.GetInt("bc_standard_dosage")
		bf.GetInt("bc_alternative_dosage")
		bf.GetInt("bc_disablebalance")
		bf.GetInt("bc_use_locations")
		bf.GetInt("time_rate")
		bf.GetLong("id_free_option")
		bf.GetInt("party_article")
		bf.GetLong("id_free_option")
		bf.GetLong("id_pua_groups")
		bf.GetInt("promo")
		bf.GetInt("one_hand_limit")
		bf.GetInt("consolidate_quantity")
		bf.GetString("consolidate_alias_name")
		bf.GetInt("hq_id")
		bf.GetBool("is_active")
		bf.GetInt("is_active_modified")
		bf.GetInt("is_active_modifier")
		bf.GetInt("rent_price_type")
		bf.GetLong("id_rental_group")
		bf.GetInt("condition_check_in_order")
		bf.GetInt("weight_required")
		bf.GetInt("daily_numeric_1")
		bf.GetInt("daily_numeric_2")
		bf.GetInt("prep_min")
		bf.GetLong("id_article_ksp")
		bf.GetInt("warn_min")
		bf.GetInt("empty_article")
		bf.GetInt("bc_debitcredit")
		bf.GetInt("prep_sec")
		bf.GetLong("id_suppliers")
		bf.GetFloat("main_price")
		bf.GetString("oman_text")
		bf.GetLong("id_age_groups")
		bf.GetInt("surcharge")
		bf.GetInt("info_data")
		bf.GetInt("pos_disabled")
		bf.GetInt("ml_name")
		bf.GetInt("ml_ks_name")
		bf.GetInt("alt_articles")
		bf.GetInt("alt_alias")
		bf.GetInt("need_prep")
		bf.GetInt("auto_onhold")
		bf.GetLong("id_ks_wf")
		bf.GetInt("ks_wf_type")
		bf.GetInt("ask_course")
		bf.GetInt("popup_info")
		bf.GetInt("allow_order_items")
		bf.GetInt("must_combined")
		bf.GetInt("block_discount")
		bf.GetInt("has_default_options")
		bf.GetInt("hht_default_setting")
		bf.GetInt("oman_default_setting")
		bf.GetLong("id_rent_periods")
		bf.GetInt("delay_separate_mins")
		bf.GetLong("id_ksc")
		bf.GetString("ml_pc_text")
		bf.GetString("ml_rm_text")
		bf.GetString("ml_oman_text")
		bf.GetInt("pos_article_type")
		bf.GetInt("single_free_option")
		bf.GetInt("ks_single_item")
		bf.GetInt("allergen")
		bf.GetInt("auto_resetcourse")
		bf.GetInt("block_transfer")
		bf.GetLong("id_size_modifier")

		bf.Release()
	}
}
func Benchmark_ReadAllArticleFields_Flat(b *testing.B) {
	bl := flatbuffers.NewBuilder(0)
	a := fillArticleFlatBuffers(bl)
	bl.Finish(a)
	bytes := bl.FinishedBytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a := GetRootAsArticle(bytes, 0)
		a.Id()
		a.ArticleNumber()
		a.Name()
		a.InternalName()
		a.ArticleManual()
		a.ArticleHash()
		a.IdCourses()
		a.IdDepartament()
		a.PcBitmap()
		a.PcColor()
		a.PcText()
		a.PcFontName()
		a.PcFontSize()
		a.PcFontAttr()
		a.PcFontColor()
		a.RmText()
		a.RmFontSize()
		a.IdPacking()
		a.IdCommission()
		a.IdPromotions()
		a.Savepoints()
		a.Quantity()
		a.Hideonhold()
		a.Barcode()
		a.TimeActive()
		a.Aftermin()
		a.Periodmin()
		a.Roundmin()
		a.IdCurrency()
		a.ControlActive()
		a.ControlTime()
		a.PluNumberVanduijnen()
		a.Sequence()
		a.RmSequence()
		a.PurchasePrice()
		a.IdVdGroup()
		a.Menu()
		a.Sensitive()
		a.SensitiveOption()
		a.DailyStock()
		a.Info()
		a.WarningLevel()
		a.FreeAfterPay()
		a.IdFoodGroup()
		a.ArticleType()
		a.IdInventoryItem()
		a.IdRecipe()
		a.IdUnitySales()
		a.CanSavepoints()
		a.ShowInKitchenScreen()
		a.DecreaseSavepoints()
		a.HhtColor()
		a.HhtFontName()
		a.HhtFontSize()
		a.HhtFontAttr()
		a.HhtFontColor()
		a.Tip()
		a.IdBecoGroup()
		a.IdBecoLocation()
		a.BcStandardDosage()
		a.BcAlternativeDosage()
		a.BcDisablebalance()
		a.BcUseLocations()
		a.TimeRate()
		a.IdFreeOption()
		a.PartyArticle()
		a.IdPuaGroups()
		a.Promo()
		a.OneHandLimit()
		a.ConsolidateQuantity()
		a.ConsolidateAliasName()
		a.HqId()
		a.IsActive()
		a.IsActiveModified()
		a.IsActiveModifier()
		a.RentPriceType()
		a.IdRentalGroup()
		a.ConditionCheckInOrder()
		a.WeightRequired()
		a.DailyNumeric1()
		a.DailyNumeric2()
		a.PrepMin()
		a.IdArticleKsp()
		a.WarnMin()
		a.EmptyArticle()
		a.BcDebitcredit()
		a.PrepSec()
		a.IdSuppliers()
		a.MainPrice()
		a.OmanText()
		a.IdAgeGroups()
		a.PosDisabled()
		a.MlName()
		a.MlKsName()
		a.AltArticles()
		a.NeedPrep()
		a.AutoOnhold()
		a.IdKsWf()
		a.KsWfType()
		a.AskCourse()
		a.AllowOrderItems()
		a.MustCombined()
		a.BlockDiscount()
		a.HasDefaultOptions()
		a.HhtDefaultSetting()
		a.OmanDefaultSetting()
		a.IdRentPeriods()
		a.DelaySeparateMins()
		a.IdKsc()
		a.MlPcText()
		a.MlRmText()
		a.MlOmanText()
		a.PosArticleType()
		a.SingleFreeOption()
		a.KsSingleItem()
		a.Allergen()
		a.AutoResetcourse()
		a.BlockTransfer()
		a.IdSizeModifier()
	}
}

func Benchmark_ReadAllArticleFields_Json(b *testing.B) {
	data, err := ioutil.ReadFile("articleData.json")
	if err != nil {
		b.Fatal(err)
	}
	jsonStr := string(data)
	dest := map[string]interface{}{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(jsonStr), &dest)
		_ = dest["id"]
		_ = dest["article_number"]
		_ = dest["name"]
		_ = dest["internal_name"]
		_ = dest["article_manual"]
		_ = dest["article_hash"]
		_ = dest["id_courses"]
		_ = dest["id_departament"]
		_ = dest["pc_bitmap"]
		_ = dest["pc_color"]
		_ = dest["pc_text"]
		_ = dest["pc_font_name"]
		_ = dest["pc_font_size"]
		_ = dest["pc_font_attr"]
		_ = dest["pc_font_color"]
		_ = dest["rm_text"]
		_ = dest["rm_font_size"]
		_ = dest["id_packing"]
		_ = dest["id_commission"]
		_ = dest["id_promotions"]
		_ = dest["savepoints"]
		_ = dest["quantity"]
		_ = dest["hideonhold"]
		_ = dest["barcode"]
		_ = dest["time_active"]
		_ = dest["aftermin"]
		_ = dest["periodmin"]
		_ = dest["roundmin"]
		_ = dest["id_currency"]
		_ = dest["control_active"]
		_ = dest["control_time"]
		_ = dest["plu_number_vanduijnen"]
		_ = dest["sequence"]
		_ = dest["rm_sequence"]
		_ = dest["purchase_price"]
		_ = dest["id_vd_group"]
		_ = dest["menu"]
		_ = dest["sensitive"]
		_ = dest["sensitive_option"]
		_ = dest["daily_stock"]
		_ = dest["info"]
		_ = dest["warning_level"]
		_ = dest["free_after_pay"]
		_ = dest["id_food_group"]
		_ = dest["article_type"]
		_ = dest["id_inventory_item"]
		_ = dest["id_recipe"]
		_ = dest["id_unity_sales"]
		_ = dest["can_savepoints"]
		_ = dest["show_in_kitchen_screen"]
		_ = dest["decrease_savepoints"]
		_ = dest["hht_color"]
		_ = dest["hht_font_name"]
		_ = dest["hht_font_size"]
		_ = dest["hht_font_attr"]
		_ = dest["hht_font_color"]
		_ = dest["tip"]
		_ = dest["id_beco_group"]
		_ = dest["id_beco_location"]
		_ = dest["bc_standard_dosage"]
		_ = dest["bc_alternative_dosage"]
		_ = dest["bc_disablebalance"]
		_ = dest["bc_use_locations"]
		_ = dest["time_rate"]
		_ = dest["id_free_option"]
		_ = dest["party_article"]
		_ = dest["id_pua_groups"]
		_ = dest["promo"]
		_ = dest["one_hand_limit"]
		_ = dest["consolidate_quantity"]
		_ = dest["consolidate_alias_name"]
		_ = dest["hq_id"]
		_ = dest["is_active"]
		_ = dest["is_active_modified"]
		_ = dest["is_active_modifier"]
		_ = dest["rent_price_type"]
		_ = dest["id_rental_group"]
		_ = dest["condition_check_in_order"]
		_ = dest["weight_required"]
		_ = dest["daily_numeric_1"]
		_ = dest["daily_numeric_2"]
		_ = dest["prep_min"]
		_ = dest["id_article_ksp"]
		_ = dest["warn_min"]
		_ = dest["empty_article"]
		_ = dest["bc_debitcredit"]
		_ = dest["prep_sec"]
		_ = dest["id_suppliers"]
		_ = dest["main_price"]
		_ = dest["oman_text"]
		_ = dest["id_age_groups"]
		_ = dest["surcharge"]
		_ = dest["info_data"]
		_ = dest["pos_disabled"]
		_ = dest["ml_name"]
		_ = dest["ml_ks_name"]
		_ = dest["alt_articles"]
		_ = dest["alt_alias"]
		_ = dest["need_prep"]
		_ = dest["auto_onhold"]
		_ = dest["id_ks_wf"]
		_ = dest["ks_wf_type"]
		_ = dest["ask_course"]
		_ = dest["popup_info"]
		_ = dest["allow_order_items"]
		_ = dest["must_combined"]
		_ = dest["block_discount"]
		_ = dest["has_default_options"]
		_ = dest["hht_default_setting"]
		_ = dest["oman_default_setting"]
		_ = dest["id_rent_periods"]
		_ = dest["delay_separate_mins"]
		_ = dest["id_ksc"]
		_ = dest["ml_pc_text"]
		_ = dest["ml_rm_text"]
		_ = dest["ml_oman_text"]
		_ = dest["pos_article_type"]
		_ = dest["single_free_option"]
		_ = dest["ks_single_item"]
		_ = dest["allergen"]
		_ = dest["auto_resetcourse"]
		_ = dest["block_transfer"]
		_ = dest["id_size_modifier"]
	}
}

func BenchmarkMap(b *testing.B) {
	scheme := map[string]interface{}{}
	for i := 0; i < 122; i++ {
		key := fmt.Sprint("field", i)
		scheme[key] = 123

	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scheme["field1"]
		_ = scheme["field2"]
	}
}

func getSimpleScheme() *dynobuffers.Scheme {
	s, _ := dynobuffers.YamlToScheme(`
name: string
price: float
quantity: int
newField: long
`)
	return s
}

func getNestedScheme() *dynobuffers.Scheme {
	s, err := dynobuffers.YamlToScheme(`
ViewMods..:
  ViewType: string
  PartitionKey: 
    Value: string
  ClusterKey: 
    Value: string
  Values: 
    field0: string
    field1: string
    field2: string
    field3: string
    field4: string
    field5: string
    field6: string
    field7: string
    field8: string
    field9: string
`)

	if err != nil {
		fmt.Print(err)
	}
	return s
}

func getNestedData() map[string]interface{} {
	view := map[string]interface{}{}

	view["viewType"] = "user"

	pkey := map[string]interface{}{}
	pkey["value"] = "user1234"

	view["partitionKey"] = pkey

	ckey := map[string]interface{}{}
	ckey["value"] = "132412341324134"

	view["clusterKey"] = ckey

	values := map[string]interface{}{
		"field0": "0[u+S=3P#)&v3Uc\"/60'z&^{17>9Po%Z%1C-06d>C-&D+0Rm0^)5!d3::570;Ri7Y'5Ow*&|.%h?=l:A#<Jk ^7!$p:6v6Ww:Gm;",
		"field1": ")\"60A?+Wg\"Ew;Tk:<v7D5$Oc*@u:-p\"T/%E/'K75_-0_%9Jc8L}\"P58Xi6N2Uo;^q(]')D%1b;W7.S;89n/; ) ~4H10[=.#b:",
		"field2": ";P9 !>-)0!; # $(+>8Wm3Cs<#h*Q/5N18A;$9v7D-19($H/$$d*I{>T5):f?@5=Aw$ :49n68\"1&65Uc;+x3Ay+[o?])(((;2b>",
		"field3": ";P3 ]7 >v/P1 Q55J'9Mq6O} &\"$P{/+|2R}#6v/4~6R9)Ky.Zw/:*3A/<Y},Yk/(>?V3%8|5:b.J{#3x+^+03 4Vg#Og?/p&W=<",
		//"field7": "35\"-H'1]};..1[!5Q7$$( Fk5C5!?.\"1>)/$%Tw\"E}$Mq\"\"r2Ia?Ni$\",<Xi#9\"5!t\"Be:U+#/f+-~51.,N{8\\{:I%5Fc*Hi1=,%",
		//"field6": "<M#)*b'U!4L?;>x#.l766,^a*@;9Qa%*`+;n(C7$\\q9=`3&|9I5$/f,S5(/f7T38]31Wg;:,(&:)(r%#|'=f3Sw#M1/G+0V7.$n>",
		//"field9": "&Co3^g.+ ,.\"-Mc;,d!/z\"(48[u+_!=b\"Fu4#8-/z0Wg&W-%.r3+l'/r3/4'^3%(()Z%)2.%\"|#Bw6Y)5!*/1`'V9!J#93*9Ea7",
		//"field8": ";5$<Ce4O%1:<.=:,8~:-|++()Ie65b4-t6Mw#2b\"6r9M5(\"h756,\\y8^k>\"~<M)(V1$Ig:. 5;44Rg6@50 $890<.d(+*8>t*Z+-",
		//"field5": "$Q3$22;=t:@k#U),5z%Yy'Ke(5r:S%?G!<Hg1Ii2Fi26p*E?8<j2\"`8W-126.D%)&v1!v:]w , $Bq/04 Wk%+$5Ia,$$)3x4^#4",
		//"field4": "9Hq?B{\"- <Qa4Y-(Ku$<60#d##p*?6&+x/..&5l:0v'T3-Cw:., '(( :4Zw9; 8Um 80=<>7L%9+l6R?1(z??x;.4$Bu2W5%Mw7",
	}

	view["values"] = values

	data := map[string]interface{}{}

	views := make([]interface{}, 1)
	views[0] = view
	data["viewMods"] = views

	return data
}
