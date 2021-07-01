/*
 * Copyright (c) 2019-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package benchmarks

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/untillpro/dynobuffers"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/linkedin/goavro"
)

func Benchmark_RW_Simple_Dyno_Typed(b *testing.B) {
	s := getSimpleScheme()
	bf := dynobuffers.NewBuffer(s)
	bf.Set("name", "cola")
	bf.Set("price", float32(0.123))
	bf.Set("quantity", int32(42))
	bytes, err := bf.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	bf.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		sum := float32(0)
		for p.Next() {
			bf := dynobuffers.ReadBuffer(bytes, s)
			price, _ := bf.GetFloat32("price")
			quantity, _ := bf.GetInt32("quantity")
			sum += price * float32(quantity)
			bf.Set("quantity", int32(3))
			if _, err := bf.ToBytes(); err != nil {
				b.Fatal(err)
			}

			bf.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_RW_Simple_Dyno_Typed_String(b *testing.B) {
	s := getSimpleScheme()
	bf := dynobuffers.NewBuffer(s)
	bf.Set("name", "cola")
	bf.Set("price", float32(0.123))
	bf.Set("quantity", int32(42))
	bytes, err := bf.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	bf.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		sum := float32(0)
		for p.Next() {
			bf := dynobuffers.ReadBuffer(bytes, s)
			price, _ := bf.GetFloat32("price")
			quantity, _ := bf.GetInt32("quantity")
			_, _ = bf.GetString("name")
			sum += price * float32(quantity)
			bf.Set("name", "new")
			if _, err := bf.ToBytes(); err != nil {
				b.Fatal(err)
			}

			bf.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_RW_Simple_Dyno_Untyped(b *testing.B) {
	s := getSimpleScheme()
	bf := dynobuffers.NewBuffer(s)
	bf.Set("name", "cola")
	bf.Set("price", float32(0.123))
	bf.Set("quantity", int32(42))
	bytes, err := bf.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	bf.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		sum := float32(0)
		for p.Next() {
			bf := dynobuffers.ReadBuffer(bytes, s)
			price := bf.Get("price").(float32) // 1 alloc here
			quantity := bf.Get("quantity").(int32)
			sum += price * float32(quantity)
			bf.Set("quantity", int32(3))
			if _, err := bf.ToBytes(); err != nil {
				b.Fatal(err)
			}
			bf.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_RW_Simple_Flat(b *testing.B) {
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
	b.RunParallel(func(p *testing.PB) {
		bl := flatbuffers.NewBuilder(0)
		sum := float32(0)
		for p.Next() {
			saleNew := GetRootAsSale(bytes, 0)
			sum += saleNew.Price() * float32(saleNew.Quantity())
			colaName := bl.CreateString("cola")
			SaleStart(bl)
			SaleAddName(bl, colaName)
			SaleAddPrice(bl, 0.123)
			SaleAddQuantity(bl, int32(1))
			sale := SaleEnd(bl)
			bl.Finish(sale)
			_ = bl.FinishedBytes()
			bl.Reset()
		}
	})
}

func Benchmark_RW_Simple_Json(b *testing.B) {
	data := map[string]interface{}{
		"name":     "cola",
		"price":    0.123,
		"quantity": int32(1),
	}
	bytes, err := json.Marshal(data)
	require.Nil(b, err)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		sum := float32(0)
		for p.Next() {
			data := map[string]interface{}{}
			if err := json.Unmarshal(bytes, &data); err != nil {
				b.Fatal(err)
			}
			price := float32(data["price"].(float64))
			sum += price * float32(data["quantity"].(float64))
			data["quantity"] = 123
			if _, err := json.Marshal(data); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Benchmark_RW_Simple_Avro(b *testing.B) {
	codec, err := goavro.NewCodec(`
		{"namespace": "unTill",
		"type": "record",
		"name": "OrderItem",
		"fields": [
			{"name": "name", "type": "string"},
			{"name": "price", "type": "float"},
			{"name": "quantity", "type": "int"}
		]}
	`)
	require.Nil(b, err)
	data := map[string]interface{}{
		"name":     string("cola"),
		"price":    float32(0.123),
		"quantity": 1,
	}
	bytes, err := codec.BinaryFromNative(nil, data)
	require.Nil(b, err)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		sum := float32(0)
		for p.Next() {
			native, _, err := codec.NativeFromBinary(bytes)
			if err != nil {
				b.Fatal(err)
			}
			decoded := native.(map[string]interface{})
			price := decoded["price"].(float32)
			sum += price * float32(decoded["quantity"].(int32))
			decoded["quantity"] = 123
			if _, err = codec.BinaryFromNative(nil, decoded); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Benchmark_RW_Article_FewFields_Dyno_Typed(b *testing.B) {
	s := getArticleSchemeDynoBuffer()
	bf := dynobuffers.NewBuffer(s)
	fillArticleDynoBuffer(bf)
	bytes, err := bf.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	bf.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		sum := float64(0)
		for p.Next() {
			bf := dynobuffers.ReadBuffer(bytes, s)
			q, _ := bf.GetInt32("quantity")
			price, _ := bf.GetFloat32("purchase_price")
			sum += float64(float32(q) * price)
			bf.Set("quantity", int32(123))
			if _, err := bf.ToBytes(); err != nil {
				b.Fatal(b, err)
			}
			bf.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}
func Benchmark_RW_Article_FewFields_Flat(b *testing.B) {
	bl := flatbuffers.NewBuilder(0)
	a := fillArticleFlatBuffers(bl)
	bl.Finish(a)
	bytes := bl.FinishedBytes()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		sum := float32(0)
		bl := flatbuffers.NewBuilder(0)
		for p.Next() {
			ar := GetRootAsArticle(bytes, 0)
			sum += ar.PurchasePrice() * float32(ar.Quantity())
			a := fillArticleFlatBuffers(bl)
			bl.Finish(a)
			_ = bl.FinishedBytes()
			bl.Reset()
		}
	})
}

func Benchmark_RW_Article_FewFields_Json(b *testing.B) {
	s := getArticleSchemeDynoBuffer()
	bf := dynobuffers.NewBuffer(s)
	fillArticleDynoBuffer(bf)
	jsonBytes := bf.ToJSON()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		sum := float64(0)
		for p.Next() {
			dest := map[string]interface{}{}
			if err := json.Unmarshal(jsonBytes, &dest); err != nil {
				b.Fatal(err)
			}
			_ = dest["id_courses"]
			q := dest["quantity"].(float64)
			price := dest["purchase_price"].(float64)
			sum += q * price
			dest["quantity"] = 123
			if _, err := json.Marshal(dest); err != nil {
				b.Fatal(err)
			}
		}
	})
	bf.Release()
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_RW_Article_FewFields_Avro(b *testing.B) {
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
	b.RunParallel(func(p *testing.PB) {
		sum := float64(0)
		for p.Next() {
			native, _, err := codec.NativeFromBinary(bytes)
			if err != nil {
				b.Fatal(err)
			}
			decoded := native.(map[string]interface{})
			price := float64(decoded["purchase_price"].(float32))
			sum += price * float64(decoded["quantity"].(int32))
			decoded["quantity"] = 123
			if _, err = codec.BinaryFromNative(nil, decoded); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Benchmark_RW_Article_AllFields_Flat(b *testing.B) {
	bl := flatbuffers.NewBuilder(0)
	a := fillArticleFlatBuffers(bl)
	bl.Finish(a)
	bytes := bl.FinishedBytes()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		bl := flatbuffers.NewBuilder(0)
		for p.Next() {
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

			ar := fillArticleFlatBuffers(bl)
			bl.Finish(ar)
			_ = bl.FinishedBytes()
			bl.Reset()
		}
	})
}

func Benchmark_RW_Article_AllFields_Json(b *testing.B) {
	data, err := ioutil.ReadFile("articleData.json")
	require.Nil(b, err)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			dest := map[string]interface{}{}
			if err := json.Unmarshal(data, &dest); err != nil {
				b.Fatal(err)
			}
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
			dest["quantity"] = 123
			if _, err := json.Marshal(dest); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Benchmark_RW_Article_AllFields_Dyno_Untyped(b *testing.B) {
	s := getArticleSchemeDynoBuffer()
	bf := dynobuffers.NewBuffer(s)
	fillArticleDynoBuffer(bf)
	bytes, err := bf.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	bf.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
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
			bf.Set("quantity", int32(123)) // to force re-encode
			if _, err := bf.ToBytes(); err != nil {
				b.Fatal(err)
			}
			bf.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_RW_Article_AllFields_Dyno_Typed(b *testing.B) {
	s := getArticleSchemeDynoBuffer()
	bf := dynobuffers.NewBuffer(s)
	fillArticleDynoBuffer(bf)
	bytes, err := bf.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	bf.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			bf := dynobuffers.ReadBuffer(bytes, s)
			bf.GetInt64("id")
			bf.GetInt32("article_number")
			bf.GetString("name")
			bf.GetString("internal_name")
			bf.GetInt32("article_manual")
			bf.GetInt32("article_hash")
			bf.GetInt64("id_courses")
			bf.GetInt64("id_departament")
			bf.GetInt32("pc_bitmap")
			bf.GetInt32("pc_color")
			bf.GetInt32("pc_text")
			bf.GetString("pc_font_name")
			bf.GetInt32("pc_font_size")
			bf.GetInt32("pc_font_attr")
			bf.GetInt32("pc_font_color")
			bf.GetString("rm_text")
			bf.GetInt32("rm_font_size")
			bf.GetInt32("id_packing")
			bf.GetInt64("id_commission")
			bf.GetInt64("id_promotions")
			bf.GetInt32("savepoints")
			bf.GetInt32("quantity")
			bf.GetInt32("hideonhold")
			bf.GetInt32("barcode")
			bf.GetInt32("time_active")
			bf.GetInt32("aftermin")
			bf.GetInt32("periodmin")
			bf.GetInt32("roundmin")
			bf.GetInt64("id_currency")
			bf.GetInt32("control_active")
			bf.GetInt32("control_time")
			bf.GetInt32("plu_number_vanduijnen")
			bf.GetInt32("sequence")
			bf.GetInt32("rm_sequence")
			bf.GetFloat32("purchase_price")
			bf.GetInt64("id_vd_group")
			bf.GetInt32("menu")
			bf.GetInt32("sensitive")
			bf.GetInt32("sensitive_option")
			bf.GetInt32("daily_stock")
			bf.GetString("info")
			bf.GetInt32("warning_level")
			bf.GetInt32("free_after_pay")
			bf.GetInt64("id_food_group")
			bf.GetInt32("article_type")
			bf.GetInt64("id_inventory_item")
			bf.GetInt64("id_recipe")
			bf.GetInt64("id_unity_sales")
			bf.GetInt32("can_savepoints")
			bf.GetBool("show_in_kitchen_screen")
			bf.GetInt32("decrease_savepoints")
			bf.GetInt32("hht_color")
			bf.GetString("hht_font_name")
			bf.GetInt32("hht_font_size")
			bf.GetInt32("hht_font_attr")
			bf.GetInt32("hht_font_color")
			bf.GetInt32("tip")
			bf.GetInt64("id_beco_group")
			bf.GetInt64("id_beco_location")
			bf.GetInt32("bc_standard_dosage")
			bf.GetInt32("bc_alternative_dosage")
			bf.GetInt32("bc_disablebalance")
			bf.GetInt32("bc_use_locations")
			bf.GetInt32("time_rate")
			bf.GetInt64("id_free_option")
			bf.GetInt32("party_article")
			bf.GetInt64("id_free_option")
			bf.GetInt64("id_pua_groups")
			bf.GetInt32("promo")
			bf.GetInt32("one_hand_limit")
			bf.GetInt32("consolidate_quantity")
			bf.GetString("consolidate_alias_name")
			bf.GetInt32("hq_id")
			bf.GetBool("is_active")
			bf.GetInt32("is_active_modified")
			bf.GetInt32("is_active_modifier")
			bf.GetInt32("rent_price_type")
			bf.GetInt64("id_rental_group")
			bf.GetInt32("condition_check_in_order")
			bf.GetInt32("weight_required")
			bf.GetInt32("daily_numeric_1")
			bf.GetInt32("daily_numeric_2")
			bf.GetInt32("prep_min")
			bf.GetInt64("id_article_ksp")
			bf.GetInt32("warn_min")
			bf.GetInt32("empty_article")
			bf.GetInt32("bc_debitcredit")
			bf.GetInt32("prep_sec")
			bf.GetInt64("id_suppliers")
			bf.GetFloat32("main_price")
			bf.GetString("oman_text")
			bf.GetInt64("id_age_groups")
			bf.GetInt32("surcharge")
			bf.GetInt32("info_data")
			bf.GetInt32("pos_disabled")
			bf.GetInt32("ml_name")
			bf.GetInt32("ml_ks_name")
			bf.GetInt32("alt_articles")
			bf.GetInt32("alt_alias")
			bf.GetInt32("need_prep")
			bf.GetInt32("auto_onhold")
			bf.GetInt64("id_ks_wf")
			bf.GetInt32("ks_wf_type")
			bf.GetInt32("ask_course")
			bf.GetInt32("popup_info")
			bf.GetInt32("allow_order_items")
			bf.GetInt32("must_combined")
			bf.GetInt32("block_discount")
			bf.GetInt32("has_default_options")
			bf.GetInt32("hht_default_setting")
			bf.GetInt32("oman_default_setting")
			bf.GetInt64("id_rent_periods")
			bf.GetInt32("delay_separate_mins")
			bf.GetInt64("id_ksc")
			bf.GetString("ml_pc_text")
			bf.GetString("ml_rm_text")
			bf.GetString("ml_oman_text")
			bf.GetInt32("pos_article_type")
			bf.GetInt32("single_free_option")
			bf.GetInt32("ks_single_item")
			bf.GetInt32("allergen")
			bf.GetInt32("auto_resetcourse")
			bf.GetInt32("block_transfer")
			bf.GetInt64("id_size_modifier")
			bf.Set("quantity", int32(123)) // to force re-encode
			if _, err := bf.ToBytes(); err != nil {
				b.Fatal(err)
			}
			bf.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}
