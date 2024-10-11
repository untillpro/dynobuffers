/*
 * Copyright (c) 2019-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

// nolint stylecheck
package benchmarks

import (
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/untillpro/dynobuffers"
)

var (

	/*
	   {"viewMods": [{"viewType": "usertable","partitionKey":{"value": "user40"},"clusterKey":{"value": "52466453699787802"},"values":{"field0":0[u+S=3P#)&v3Uc\"/60'z&^{17>9Po%Z%1C-06d>C-&D+0Rm0^)5!d3::570;Ri7Y'5Ow*&|.%h?=l:A#<Jk ^7!$p:6v6Ww:Gm;","field1": ")\"60A?+Wg\"Ew;Tk:<v7D5$Oc*@u:-p\"T/%E/'K75_-0_%9Jc8L}\"P58Xi6N2Uo;^q(]')D%1b;W7.S;89n/; ) ~4H10[=.#b:","field2": ";P9 !>-)0!; # $(+>8Wm3Cs<#h*Q/5N18A;$9v7D-19($H/$$d*I{>T5):f?@5=Aw$ :49n68\"1&65Uc;+x3Ay+[o?])(((;2b>","field3": ";P3 ]7 >v/P1 Q55J'9Mq6O} &\"$P{/+|2R}#6v/4~6R9)Ky.Zw/:*3A/<Y},Yk/(>?V3%8|5:b.J{#3x+^+03 4Vg#Og?/p&W=<","field4": "9Hq?B{\"- <Qa4Y-(Ku$<60#d##p*?6&+x/..&5l:0v'T3-Cw:., '(( :4Zw9; 8Um 80=<>7L%9+l6R?1(z??x;.4$Bu2W5%Mw7","field5": "$Q3$22;=t:@k#U),5z%Yy'Ke(5r:S%?G!<Hg1Ii2Fi26p*E?8<j2\"8W-126.D%)&v1!v:]w , $Bq/04 Wk%+$5Ia,$$)3x4^#4","field6": "<M#)*b'U!4L?;>x#.l766,^a*@;9Qa%*+;n(C7$\\q9=3&|9I5$/f,S5(/f7T38]31Wg;:,(&:)(r%#|'=f3Sw#M1/G+0V7.$n>","field7": "35\"-H'1]};..1[!5Q7$$( Fk5C5!?.\"1>)/$%Tw\"E}$Mq\"\"r2Ia?Ni$\",<Xi#9\"5!t\"Be:U+#/f+-~51.,N{8\\{:I%5Fc*Hi1=,%","field8": ";5$<Ce4O%1:<.=:,8~:-|++()Ie65b4-t6Mw#2b\"6r9M5(\"h756,\\y8^k>\"~<M)(V1$Ig:. 5;44Rg6@50 $890<.d(+*8>t*Z+-","field9": "&Co3^g.+ ,.\"-Mc;,d!/z\"(48[u+_!=b\"Fu4#8-/z0Wg&W-%.r3+l'/r3/4'^3%(()Z%)2.%\"|#Bw6Y)5!*1'V9!J#93*9Ea7"}}]}
	*/

	testData = []byte(`
{
	"viewMods": [{
		"viewType": "usertable",
		"partitionKey": {
			"value": "user40"
		},
		"clusterKey": {
			"value": "52466453699787802"
		},
		"values": {
			"field0": "0[u+S=3P#)&v3Uc\"/60'z&^{17>9Po%Z%1C-06d>C-&D+0Rm0^)5!d3::570;Ri7Y'5Ow*&|.%h?=l:A#<Jk ^7!$p:6v6Ww:Gm;",
			"field1": ")\"60A?+Wg\"Ew;Tk:<v7D5$Oc*@u:-p\"T/%E/'K75_-0_%9Jc8L}\"P58Xi6N2Uo;^q(]')D%1b;W7.S;89n/; ) ~4H10[=.#b:",
			"field2": ";P9 !>-)0!; # $(+>8Wm3Cs<#h*Q/5N18A;$9v7D-19($H/$$d*I{>T5):f?@5=Aw$ :49n68\"1&65Uc;+x3Ay+[o?])(((;2b>",
			"field3": ";P3 ]7 >v/P1 Q55J'9Mq6O} &\"$P{/+|2R}#6v/4~6R9)Ky.Zw/:*3A/<Y},Yk/(>?V3%8|5:b.J{#3x+^+03 4Vg#Og?/p&W=<",
			"field4": "9Hq?B{\"- <Qa4Y-(Ku$<60#d##p*?6&+x/..&5l:0v'T3-Cw:., '(( :4Zw9; 8Um 80=<>7L%9+l6R?1(z??x;.4$Bu2W5%Mw7",
			"field5": "$Q3$22;=t:@k#U),5z%Yy'Ke(5r:S%?G!<Hg1Ii2Fi26p*E?8<j2\"8W-126.D%)&v1!v:]w , $Bq/04 Wk%+$5Ia,$$)3x4^#4",
			"field6": "<M#)*b'U!4L?;>x#.l766,^a*@;9Qa%*+;n(C7$\\q9=3&|9I5$/f,S5(/f7T38]31Wg;:,(&:)(r%#|'=f3Sw#M1/G+0V7.$n>",
			"field7": "35\"-H'1]};..1[!5Q7$$( Fk5C5!?.\"1>)/$%Tw\"E}$Mq\"\"r2Ia?Ni$\",<Xi#9\"5!t\"Be:U+#/f+-~51.,N{8\\{:I%5Fc*Hi1=,%",
			"field8": ";5$<Ce4O%1:<.=:,8~:-|++()Ie65b4-t6Mw#2b\"6r9M5(\"h756,\\y8^k>\"~<M)(V1$Ig:. 5;44Rg6@50 $890<.d(+*8>t*Z+-",
			"field9": "&Co3^g.+ ,.\"-Mc;,d!/z\"(48[u+_!=b\"Fu4#8-/z0Wg&W-%.r3+l'/r3/4'^3%(()Z%)2.%\"|#Bw6Y)5!*1'V9!J#93*9Ea7"
		}
	}]
}`)
	pbillYaml string = `
Id: int64
Id_bill: int64
Id_untill_users: int64
number: int32
failurednumber: int32
suffix: string
pdatetime: string
id_sales_area: int64
pcname: string
service_charge: float64
real_datetime: string
tips: float32
id_clients: int64
pbill_index: int32
split_parts: float64
id_real_untill_user: int64
reopen_type: byte
external_id: string
covers: int32
hht: string
super_dx: int64
c_tips: float32
id_currency: int64
bill:
  Id: int64
  Tableno: int32
  Id_untill_users: int64
  Table_part: string
  id_courses: int64
  id_clients: int64
  name: string
  Proforma: byte
  modified: string
  open_datetime: string
  close_datetime: string
  number: int32
  failurednumber: int32
  suffix: string
  pbill_number: int32
  pbill_failurednumber: int32
  pbill_suffix: string
  hc_foliosequence: int32
  hc_folionumber: string
  tip: float32
  qty_persons: int32
  isdirty: byte
  reservationid: string
  id_alter_user: int64
  service_charge: float64
  number_of_covers: int32
  id_user_proforma: int64
  bill_type: byte
  locker: int32
  id_time_article: int64
  timer_start: string
  timer_stop: string
  isactive: byte
  table_name: string
  group_vat_level: byte
  comments: string
  id_cardprice: int64
  discount: float64
  discount_value: float32
  id_discount_reasons: int64
  hc_roomnumber: string
  ignore_auto_sc: byte
  extra_fields..: byte
  id_bo_service_charge: int64
  free_comments: string
  id_t2o_groups: int64
  service_tax: float32
  sc_plan..: byte
  client_phone: string
  age: string
  description..: byte
  sdescription: string
  vars..: byte
  take_away: int32
  fiscal_number: int32
  fiscal_failurednumber: int32
  fiscal_suffix: string
  id_order_type: int64
  not_paid: float32
  total: float32
bill_reprints..:
  Id: int64
  Id_pbill: int64
  datetime: string
  id_untill_users: int64
bill_split_rest..:
  Id: int64
  id_bill: int64
  is_active: byte
  split_data..: byte
  datetime: string
  id_pbill: int64
complete_meal..:
  Id: int64
  id_pbill: int64
  id_untill_users: int64
  cm_datetime: string
  cm_pcname: string
  cm_quantity: int32
  cm_price: float32
neg_pbill..:
  Id: int64
  Id_pbill: int64
  R_type: byte
  pdatetime: string
  id_pbill_close: int64
open_discounts..:
  Id: int64
  Id_pbill: int64
  amount: float32
  vat_percent: float32
  vat: float64
pbill_balance..:
  Id: int64
  id_pbill: int64
  id_clients: int64
  id_accounts: int64
  pre_balance: float32
  pre_balance_common: float32
pbill_cleancash_info..:
  Id: int64
  id_pbill: int64
  clean_cash_trnumber: int64
  clean_cash_signature: string
  manufacturing_code: string
  control_timestamp: string
  ticket_counter: string
  vsc_id: string
  plu_data_hash: string
  id_bill: int64
  ccdatetime: string
  kind: byte
  production_nr: string
  extra1: string
  extra2: string
pbill_email..:
  Id: int64
  id_pbill: int64
  pemail_email: string
  pemail_body..: byte
  pemail_dt: string
  pemail_state: byte
pbill_item..:
  Id: int64
  Id_pbill: int64
  Id_untill_users: int64
  Tableno: int32
  Rowbeg: byte
  Kind: byte
  Quantity: int32
  id_articles: int64
  Price: float32
  text: string
  id_prices: int64
  with_without: int32
  id_courses: int64
  pdatetime: string
  purchase_price: float32
  vat_sign: string
  vat: float64
  id_menu: int64
  menu_quantity: int32
  original_price: float32
  id_discount_reasons: int64
  discount_type: int32
  chair_number: int32
  separated: byte
  negative: byte
  vat_percent: float32
  chair_name: string
  id_option_article: int64
  start_delay_minutes: int32
  pua: byte
  no_price: byte
  orig_pbill_id: int64
  parts_id: int64
  full_paid: byte
  weight: float32
  splitted_coef: float64
  split_id: int64
  free_comments: string
  id_smartcards: int64
  has_allergens: byte
  id_article_options: int64
  order_item_allergens..:
    Id: int64
    id_order_item: int64
    id_menu_item: int64
    id_allergens: int64
    text: string
    id_pbill_item: int64
  order_item_extra..:
    Id: int64
    id_order_item: int64
    id_menu_item: int64
    id_pbill_item: int64
    id_article_barcodes: int64
  order_item_sizes..:
    Id: int64
    id_order_item: int64
    id_menu_item: int64
    id_pbill_item: int64
    id_articles: int64
    price: float32
    original_price: float32
    vat: float64
    id_size_modifier_item: int64
  pbill_hash..:
    Id: int64
    Id_pbill_item: int64
    Id_untill_users: int64
  pbill_item_bookp..:
    Id: int64
    id_pbill_item: int64
    id_bookkeeping_turnover: int64
    id_bookkeeping_vat: int64
pbill_item_refund..:
  Id: int64
  Id_pbill: int64
  Id_untill_users: int64
  Tableno: int32
  Rowbeg: byte
  Kind: byte
  Quantity: int32
  id_articles: int64
  id_department: int64
  id_food_group: int64
  id_category: int64
  Price: float32
  text: string
  id_prices: int64
  with_without: int32
  id_courses: int64
  pdatetime: string
  purchase_price: float32
  vat_sign: string
  vat: float64
  id_menu: int64
  menu_quantity: int32
  ref_type: byte
pbill_payments..:
  Id: int64
  Id_pbill: int64
  Id_payments: int64
  Price: float32
  customer_amount: float32
  c_price: float32
  c_customer_amount: float32
  vat: float32
  id_driver_discount: int64
  accounts_payments..:
    Id: int64
    datetime: string
    Id_clients: int64
    Id_untill_users: int64
    tableno: int32
    tablepart: string
    number: int32
    suffix: string
    name: string
    tr_datetime: string
    bill_number: string
    parts_count: byte
    part_index: byte
    discount: float64
    id_sales_area: int64
    total: float32
    account_type: byte
    payer_name: string
    id_pbill_payments: int64
  account_payment_data..:
    Id: int64
    id_pbill_payments: int64
    account_type: byte
  payments_hash..:
    Id: int64
    Id_payments: int64
    Id_untill_users: int64
  pbill_card_payments_info..:
    Id: int64
    Id_pbill_payments: int64
    masked_pan: string
    pan: string
    expire: string
    auth_number: string
    acquier_code: string
    field1: string
    field2: string
    field3: string
    field4: string
    field5: string
    track2data: string
    loyaltyamount: int64
    controlno: string
    chargetype: int32
    cardname: string
    cardnoentertype: int32
    terminalid: string
    aid: string
    tvr: string
    tsi: string
    account_type: string
    customer_receipt: string
    merchant_receipt: string
    tip_approvement: byte
    prepaid_pan: string
    misc_data..: byte
    field6: string
    field7: string
    field8: string
    field9: string
    field10: string
    merchant_id: string
    cashback_amount: int64
    field11: string
    field12: string
    field13: string
    field14: string
    field15: string
  pbill_payer..:
    Id: int64
    id_pbill_payments: int64
    payer_name: string
  pbill_payments_bookp..:
    Id: int64
    id_pbill_payments: int64
    id_bookkeeping: int64
  pbill_zapper_payments_info..:
    Id: int64
    Id_pbill_payments: int64
    Id_zapper_proformas: int64
    zapper_id: string
    misc_data..: byte
  psp_tips..:
    Id: int64
    Id_pbill: int64
    amount: float32
    id_pbill_payments: int64
  smartcards_turnover..:
    Id: int64
    datetime: string
    Id_smartcards: int64
    Id_untill_users: int64
    kind: byte
    amount: float32
    p_amount: float32
    id_pbill_payments: int64
    id_payments: int64
    is_initial_value: byte
    card_data..: byte
    deposit_number: int32
    deposit_suffix: string
    pcname: string
    new_balance: float32
  voucher_payments..:
    Id: int64
    id_vouchers: int64
    id_pbill_payments: int64
    used_dt: string
  voucher_pbill_payments..:
    Id: int64
    id_voucher_pbill: int64
    price: float32
    id_vouchers: int64
    id_payments: int64
    id_pbill_payments: int64
pbill_return..:
  Id: int64
  id_pbill: int64
  ref_type: byte
  id_void_reason: int64
  void_text: string
psp_tips..:
  Id: int64
  Id_pbill: int64
  amount: float32
  id_pbill_payments: int64
sold_articles..:
  Id: int64
  id_pbill: int64
  id_order_item: int64
  quantity: int32
  parts_id: int64
  sa_coef: float64
  split_id: int64
sold_articles_wm..:
  Id: int64
  id_order_item: int64
  id_pbill: int64
  quantity: int32
  sa_coef: float64
stock_doc..:
  Id: int64
  id_orders: int64
  id_pbill: int64
  id_stock_invoice: int64
  id_stock_adjustment: int64
  id_stock_cost_correction: int64
stock_pbill_queue..:
  Id: int64
  Id_pbill: int64
  pdatetime: string
`
)

func fillArticleFlatBuffers(bl *flatbuffers.Builder) flatbuffers.UOffsetT {
	name := bl.CreateString("str")
	interName := bl.CreateString("str")
	pcText := bl.CreateString("str")
	fontName := bl.CreateString("str")
	omanText := bl.CreateString("str")
	pcBitmap := bl.CreateString("str")
	rmText := bl.CreateString("str")
	barcode := bl.CreateString("str")
	info := bl.CreateString("str")
	hhtFontName := bl.CreateString("str")
	aliasName := bl.CreateString("str")
	hqID := bl.CreateString("str")
	isActiveModifier := bl.CreateString("str")
	mlName := bl.CreateString("str")
	mlKSName := bl.CreateString("str")
	mlPCText := bl.CreateString("str")
	mlRMText := bl.CreateString("str")
	mlOmanText := bl.CreateString("str")

	ArticleStart(bl)
	ArticleAddId(bl, 1)
	ArticleAddArticleNumber(bl, 1)
	ArticleAddName(bl, name)
	ArticleAddInternalName(bl, interName)
	ArticleAddArticleManual(bl, true)
	ArticleAddArticleHash(bl, true)
	ArticleAddIdCourses(bl, 1)
	ArticleAddIdDepartament(bl, 1)
	ArticleAddPcBitmap(bl, pcBitmap)
	ArticleAddPcColor(bl, 1)
	ArticleAddPcText(bl, pcText)
	ArticleAddPcFontName(bl, fontName)
	ArticleAddPcFontSize(bl, 1)
	ArticleAddPcFontAttr(bl, 1)
	ArticleAddPcFontColor(bl, 1)
	ArticleAddRmText(bl, rmText)
	ArticleAddRmFontSize(bl, 1)
	ArticleAddIdPacking(bl, 1)
	ArticleAddIdCommission(bl, 1)
	ArticleAddIdPromotions(bl, 1)
	ArticleAddSavepoints(bl, 1)
	ArticleAddQuantity(bl, 1)
	ArticleAddHideonhold(bl, true)
	ArticleAddBarcode(bl, barcode)
	ArticleAddTimeActive(bl, true)
	ArticleAddAftermin(bl, 1)
	ArticleAddPeriodmin(bl, 1)
	ArticleAddRoundmin(bl, 1)
	ArticleAddIdCurrency(bl, 1)
	ArticleAddControlActive(bl, true)
	ArticleAddControlTime(bl, 1)
	ArticleAddPluNumberVanduijnen(bl, 1)
	ArticleAddSequence(bl, 1)
	ArticleAddRmSequence(bl, 1)
	ArticleAddPurchasePrice(bl, 1)
	ArticleAddIdVdGroup(bl, 1)
	ArticleAddMenu(bl, true)
	ArticleAddSensitive(bl, true)
	ArticleAddSensitiveOption(bl, true)
	ArticleAddDailyStock(bl, 1)
	ArticleAddInfo(bl, info)
	ArticleAddWarningLevel(bl, 1)
	ArticleAddFreeAfterPay(bl, 1)
	ArticleAddIdFoodGroup(bl, 1)
	ArticleAddArticleType(bl, true)
	ArticleAddIdInventoryItem(bl, 1)
	ArticleAddIdRecipe(bl, 1)
	ArticleAddIdUnitySales(bl, 1)
	ArticleAddCanSavepoints(bl, true)
	ArticleAddShowInKitchenScreen(bl, true)
	ArticleAddDecreaseSavepoints(bl, 1)
	ArticleAddHhtColor(bl, 1)
	ArticleAddHhtFontName(bl, hhtFontName)
	ArticleAddHhtFontSize(bl, 1)
	ArticleAddHhtFontAttr(bl, 1)
	ArticleAddHhtFontColor(bl, 1)
	ArticleAddTip(bl, true)
	ArticleAddIdBecoGroup(bl, 1)
	ArticleAddIdBecoLocation(bl, 1)
	ArticleAddBcStandardDosage(bl, 1)
	ArticleAddBcAlternativeDosage(bl, 1)
	ArticleAddBcDisablebalance(bl, true)
	ArticleAddBcUseLocations(bl, true)
	ArticleAddTimeRate(bl, 1)
	ArticleAddIdFreeOption(bl, 1)
	ArticleAddPartyArticle(bl, true)
	ArticleAddIdPuaGroups(bl, 1)
	ArticleAddPromo(bl, true)
	ArticleAddOneHandLimit(bl, 1)
	ArticleAddConsolidateQuantity(bl, 1)
	ArticleAddConsolidateAliasName(bl, aliasName)
	ArticleAddHqId(bl, hqID)
	ArticleAddIsActive(bl, true)
	ArticleAddIsActiveModified(bl, 1)
	ArticleAddIsActiveModifier(bl, isActiveModifier)
	ArticleAddRentPriceType(bl, true)
	ArticleAddIdRentalGroup(bl, 1)
	ArticleAddConditionCheckInOrder(bl, true)
	ArticleAddWeightRequired(bl, true)
	ArticleAddDailyNumeric1(bl, 1)
	ArticleAddDailyNumeric2(bl, 1)
	ArticleAddPrepMin(bl, 1)
	ArticleAddIdArticleKsp(bl, 1)
	ArticleAddWarnMin(bl, 1)
	ArticleAddEmptyArticle(bl, true)
	ArticleAddBcDebitcredit(bl, true)
	ArticleAddPrepSec(bl, 1)
	ArticleAddIdSuppliers(bl, 1)
	ArticleAddMainPrice(bl, true)
	ArticleAddOmanText(bl, omanText)
	ArticleAddIdAgeGroups(bl, 1)
	ArticleAddSurcharge(bl, true)
	ArticleAddInfoData(bl, 1)
	ArticleAddPosDisabled(bl, true)
	ArticleAddMlName(bl, mlName)
	ArticleAddMlKsName(bl, mlKSName)
	ArticleAddAltArticles(bl, 1)
	ArticleAddAltAlias(bl, 1)
	ArticleAddNeedPrep(bl, true)
	ArticleAddAutoOnhold(bl, true)
	ArticleAddIdKsWf(bl, 1)
	ArticleAddKsWfType(bl, 1)
	ArticleAddAskCourse(bl, true)
	ArticleAddPopupInfo(bl, 1)
	ArticleAddAllowOrderItems(bl, true)
	ArticleAddMustCombined(bl, true)
	ArticleAddBlockDiscount(bl, true)
	ArticleAddHasDefaultOptions(bl, true)
	ArticleAddHhtDefaultSetting(bl, true)
	ArticleAddOmanDefaultSetting(bl, true)
	ArticleAddIdRentPeriods(bl, 1)
	ArticleAddDelaySeparateMins(bl, 1)
	ArticleAddIdKsc(bl, 1)
	ArticleAddMlPcText(bl, mlPCText)
	ArticleAddMlRmText(bl, mlRMText)
	ArticleAddMlOmanText(bl, mlOmanText)
	ArticleAddPosArticleType(bl, true)
	ArticleAddSingleFreeOption(bl, true)
	ArticleAddKsSingleItem(bl, true)
	ArticleAddAllergen(bl, true)
	ArticleAddAutoResetcourse(bl, true)
	ArticleAddBlockTransfer(bl, true)
	ArticleAddIdSizeModifier(bl, 1)
	return ArticleEnd(bl)
}

func fillArticleDynoBuffer(bf *dynobuffers.Buffer) {
	bf.Set("id", int64(2))
	bf.Set("article_number", int32(1))
	bf.Set("name", "str")
	bf.Set("internal_name", "str")
	bf.Set("article_manual", true)
	bf.Set("article_hash", true)
	bf.Set("id_courses", int64(2))
	bf.Set("id_departament", int64(2))
	bf.Set("pc_bitmap", "str") // blob
	bf.Set("pc_color", int32(1))
	bf.Set("pc_text", "str")
	bf.Set("pc_font_name", "str")
	bf.Set("pc_font_size", int32(1))
	bf.Set("pc_font_attr", int32(1))
	bf.Set("pc_font_color", int32(1))
	bf.Set("rm_text", "str")
	bf.Set("rm_font_size", int32(1))
	bf.Set("id_packing", int64(2))
	bf.Set("id_commission", int64(2))
	bf.Set("id_promotions", int64(2))
	bf.Set("savepoints", int32(1))
	bf.Set("quantity", int32(1))
	bf.Set("hideonhold", true)
	bf.Set("barcode", "str")
	bf.Set("time_active", true)
	bf.Set("aftermin", int32(1))
	bf.Set("periodmin", int32(1))
	bf.Set("roundmin", int32(1))
	bf.Set("id_currency", int64(2))
	bf.Set("control_active", true)
	bf.Set("control_time", int32(1))
	bf.Set("plu_number_vanduijnen", int32(1))
	bf.Set("sequence", int32(1))
	bf.Set("rm_sequence", int32(1))
	bf.Set("purchase_price", float32(0.123))
	bf.Set("id_vd_group", int64(2))
	bf.Set("menu", true)
	bf.Set("sensitive", true)
	bf.Set("sensitive_option", true)
	bf.Set("daily_stock", int32(1))
	bf.Set("info", "str")
	bf.Set("warning_level", int32(1))
	bf.Set("free_after_pay", int32(1))
	bf.Set("id_food_group", int64(2))
	bf.Set("article_type", byte(3))
	bf.Set("id_inventory_item", int64(2))
	bf.Set("id_recipe", int64(2))
	bf.Set("id_unity_sales", int64(2))
	bf.Set("can_savepoints", true)
	bf.Set("show_in_kitchen_screen", true)
	bf.Set("decrease_savepoints", int32(1))
	bf.Set("hht_color", int32(1))
	bf.Set("hht_font_name", "str")
	bf.Set("hht_font_size", int32(1))
	bf.Set("hht_font_attr", int32(1))
	bf.Set("hht_font_color", int32(1))
	bf.Set("tip", true)
	bf.Set("id_beco_group", int64(2))
	bf.Set("id_beco_location", int64(2))
	bf.Set("bc_standard_dosage", int32(1))
	bf.Set("bc_alternative_dosage", int32(1))
	bf.Set("bc_disablebalance", true)
	bf.Set("bc_use_locations", true)
	bf.Set("time_rate", float32(0.123))
	bf.Set("id_free_option", int64(2))
	bf.Set("party_article", true)
	bf.Set("id_pua_groups", int64(2))
	bf.Set("promo", true)
	bf.Set("one_hand_limit", int32(1))
	bf.Set("consolidate_quantity", int32(1))
	bf.Set("consolidate_alias_name", "str")
	bf.Set("hq_id", "str")
	bf.Set("is_active", true)
	bf.Set("is_active_modified", int32(1)) // timestamp
	bf.Set("is_active_modifier", "str")
	bf.Set("rent_price_type", true)
	bf.Set("id_rental_group", int64(2))
	bf.Set("condition_check_in_order", true)
	bf.Set("weight_required", true)
	bf.Set("daily_numeric_1", float32(0.123))
	bf.Set("daily_numeric_2", float32(0.123))
	bf.Set("prep_min", int32(1))
	bf.Set("id_article_ksp", int64(2))
	bf.Set("warn_min", int32(1))
	bf.Set("empty_article", true)
	bf.Set("bc_debitcredit", true)
	bf.Set("prep_sec", int32(1))
	bf.Set("id_suppliers", int64(2))
	bf.Set("main_price", true)
	bf.Set("oman_text", "str")
	bf.Set("id_age_groups", int64(2))
	bf.Set("surcharge", true)
	bf.Set("info_data", "str") //blob
	bf.Set("pos_disabled", true)
	bf.Set("ml_name", "str")    // blob
	bf.Set("ml_ks_name", "str") // blob
	bf.Set("alt_articles", int64(2))
	bf.Set("alt_alias", "str")
	bf.Set("need_prep", true)
	bf.Set("auto_onhold", true)
	bf.Set("id_ks_wf", int64(2))
	bf.Set("ks_wf_type", int32(1))
	bf.Set("ask_course", true)
	bf.Set("popup_info", "str")
	bf.Set("allow_order_items", true)
	bf.Set("must_combined", true)
	bf.Set("block_discount", true)
	bf.Set("has_default_options", true)
	bf.Set("hht_default_setting", true)
	bf.Set("oman_default_setting", true)
	bf.Set("id_rent_periods", int64(2))
	bf.Set("delay_separate_mins", int32(1))
	bf.Set("id_ksc", int64(2))
	bf.Set("ml_pc_text", "str")   // blob
	bf.Set("ml_rm_text", "str")   // blob
	bf.Set("ml_oman_text", "str") // blob
	bf.Set("pos_article_type", true)
	bf.Set("single_free_option", true)
	bf.Set("ks_single_item", true)
	bf.Set("allergen", true)
	bf.Set("auto_resetcourse", true)
	bf.Set("block_transfer", true)
	bf.Set("id_size_modifier", int64(2))
}

func getArticleSchemeDynoBuffer() *dynobuffers.Scheme {
	s := dynobuffers.NewScheme()
	s.AddField("id", dynobuffers.FieldTypeInt64, false)
	s.AddField("article_number", dynobuffers.FieldTypeInt32, false)
	s.AddField("name", dynobuffers.FieldTypeString, false)
	s.AddField("internal_name", dynobuffers.FieldTypeString, false)
	s.AddField("article_manual", dynobuffers.FieldTypeBool, false)
	s.AddField("article_hash", dynobuffers.FieldTypeBool, false)
	s.AddField("id_courses", dynobuffers.FieldTypeInt64, false)
	s.AddField("id_departament", dynobuffers.FieldTypeInt64, false)
	s.AddField("pc_bitmap", dynobuffers.FieldTypeString, false) // blob
	s.AddField("pc_color", dynobuffers.FieldTypeInt32, false)
	s.AddField("pc_text", dynobuffers.FieldTypeString, false)
	s.AddField("pc_font_name", dynobuffers.FieldTypeString, false)
	s.AddField("pc_font_size", dynobuffers.FieldTypeInt32, false)
	s.AddField("pc_font_attr", dynobuffers.FieldTypeInt32, false)
	s.AddField("pc_font_color", dynobuffers.FieldTypeInt32, false)
	s.AddField("rm_text", dynobuffers.FieldTypeString, false)
	s.AddField("rm_font_size", dynobuffers.FieldTypeInt32, false)
	s.AddField("id_packing", dynobuffers.FieldTypeInt64, false)
	s.AddField("id_commission", dynobuffers.FieldTypeInt64, false)
	s.AddField("id_promotions", dynobuffers.FieldTypeInt64, false)
	s.AddField("savepoints", dynobuffers.FieldTypeInt32, false)
	s.AddField("quantity", dynobuffers.FieldTypeInt32, false)
	s.AddField("hideonhold", dynobuffers.FieldTypeBool, false)
	s.AddField("barcode", dynobuffers.FieldTypeString, false)
	s.AddField("time_active", dynobuffers.FieldTypeBool, false)
	s.AddField("aftermin", dynobuffers.FieldTypeInt32, false)
	s.AddField("periodmin", dynobuffers.FieldTypeInt32, false)
	s.AddField("roundmin", dynobuffers.FieldTypeInt32, false)
	s.AddField("id_currency", dynobuffers.FieldTypeInt64, false)
	s.AddField("control_active", dynobuffers.FieldTypeBool, false)
	s.AddField("control_time", dynobuffers.FieldTypeInt32, false)
	s.AddField("plu_number_vanduijnen", dynobuffers.FieldTypeInt32, false)
	s.AddField("sequence", dynobuffers.FieldTypeInt32, false)
	s.AddField("rm_sequence", dynobuffers.FieldTypeInt32, false)
	s.AddField("purchase_price", dynobuffers.FieldTypeFloat32, false)
	s.AddField("id_vd_group", dynobuffers.FieldTypeInt64, false)
	s.AddField("menu", dynobuffers.FieldTypeBool, false)
	s.AddField("sensitive", dynobuffers.FieldTypeBool, false)
	s.AddField("sensitive_option", dynobuffers.FieldTypeBool, false)
	s.AddField("daily_stock", dynobuffers.FieldTypeInt32, false)
	s.AddField("info", dynobuffers.FieldTypeString, false)
	s.AddField("warning_level", dynobuffers.FieldTypeInt32, false)
	s.AddField("free_after_pay", dynobuffers.FieldTypeInt32, false)
	s.AddField("id_food_group", dynobuffers.FieldTypeInt64, false)
	s.AddField("article_type", dynobuffers.FieldTypeByte, false)
	s.AddField("id_inventory_item", dynobuffers.FieldTypeInt64, false)
	s.AddField("id_recipe", dynobuffers.FieldTypeInt64, false)
	s.AddField("id_unity_sales", dynobuffers.FieldTypeInt64, false)
	s.AddField("can_savepoints", dynobuffers.FieldTypeBool, false)
	s.AddField("show_in_kitchen_screen", dynobuffers.FieldTypeBool, false)
	s.AddField("decrease_savepoints", dynobuffers.FieldTypeInt32, false)
	s.AddField("hht_color", dynobuffers.FieldTypeInt32, false)
	s.AddField("hht_font_name", dynobuffers.FieldTypeString, false)
	s.AddField("hht_font_size", dynobuffers.FieldTypeInt32, false)
	s.AddField("hht_font_attr", dynobuffers.FieldTypeInt32, false)
	s.AddField("hht_font_color", dynobuffers.FieldTypeInt32, false)
	s.AddField("tip", dynobuffers.FieldTypeBool, false)
	s.AddField("id_beco_group", dynobuffers.FieldTypeInt64, false)
	s.AddField("id_beco_location", dynobuffers.FieldTypeInt64, false)
	s.AddField("bc_standard_dosage", dynobuffers.FieldTypeInt32, false)
	s.AddField("bc_alternative_dosage", dynobuffers.FieldTypeInt32, false)
	s.AddField("bc_disablebalance", dynobuffers.FieldTypeBool, false)
	s.AddField("bc_use_locations", dynobuffers.FieldTypeBool, false)
	s.AddField("time_rate", dynobuffers.FieldTypeFloat32, false)
	s.AddField("id_free_option", dynobuffers.FieldTypeInt64, false)
	s.AddField("party_article", dynobuffers.FieldTypeBool, false)
	s.AddField("id_pua_groups", dynobuffers.FieldTypeInt64, false)
	s.AddField("promo", dynobuffers.FieldTypeBool, false)
	s.AddField("one_hand_limit", dynobuffers.FieldTypeInt32, false)
	s.AddField("consolidate_quantity", dynobuffers.FieldTypeInt32, false)
	s.AddField("consolidate_alias_name", dynobuffers.FieldTypeString, false)
	s.AddField("hq_id", dynobuffers.FieldTypeString, false)
	s.AddField("is_active", dynobuffers.FieldTypeBool, false)
	s.AddField("is_active_modified", dynobuffers.FieldTypeInt32, false) // timestamp
	s.AddField("is_active_modifier", dynobuffers.FieldTypeString, false)
	s.AddField("rent_price_type", dynobuffers.FieldTypeBool, false)
	s.AddField("id_rental_group", dynobuffers.FieldTypeInt64, false)
	s.AddField("condition_check_in_order", dynobuffers.FieldTypeBool, false)
	s.AddField("weight_required", dynobuffers.FieldTypeBool, false)
	s.AddField("daily_numeric_1", dynobuffers.FieldTypeFloat32, false)
	s.AddField("daily_numeric_2", dynobuffers.FieldTypeFloat32, false)
	s.AddField("prep_min", dynobuffers.FieldTypeInt32, false)
	s.AddField("id_article_ksp", dynobuffers.FieldTypeInt64, false)
	s.AddField("warn_min", dynobuffers.FieldTypeInt32, false)
	s.AddField("empty_article", dynobuffers.FieldTypeBool, false)
	s.AddField("bc_debitcredit", dynobuffers.FieldTypeBool, false)
	s.AddField("prep_sec", dynobuffers.FieldTypeInt32, false)
	s.AddField("id_suppliers", dynobuffers.FieldTypeInt64, false)
	s.AddField("main_price", dynobuffers.FieldTypeBool, false)
	s.AddField("oman_text", dynobuffers.FieldTypeString, false)
	s.AddField("id_age_groups", dynobuffers.FieldTypeInt64, false)
	s.AddField("surcharge", dynobuffers.FieldTypeBool, false)
	s.AddField("info_data", dynobuffers.FieldTypeString, false) //blob
	s.AddField("pos_disabled", dynobuffers.FieldTypeBool, false)
	s.AddField("ml_name", dynobuffers.FieldTypeString, false)    // blob
	s.AddField("ml_ks_name", dynobuffers.FieldTypeString, false) // blob
	s.AddField("alt_articles", dynobuffers.FieldTypeInt64, false)
	s.AddField("alt_alias", dynobuffers.FieldTypeString, false)
	s.AddField("need_prep", dynobuffers.FieldTypeBool, false)
	s.AddField("auto_onhold", dynobuffers.FieldTypeBool, false)
	s.AddField("id_ks_wf", dynobuffers.FieldTypeInt64, false)
	s.AddField("ks_wf_type", dynobuffers.FieldTypeInt32, false)
	s.AddField("ask_course", dynobuffers.FieldTypeBool, false)
	s.AddField("popup_info", dynobuffers.FieldTypeString, false)
	s.AddField("allow_order_items", dynobuffers.FieldTypeBool, false)
	s.AddField("must_combined", dynobuffers.FieldTypeBool, false)
	s.AddField("block_discount", dynobuffers.FieldTypeBool, false)
	s.AddField("has_default_options", dynobuffers.FieldTypeBool, false)
	s.AddField("hht_default_setting", dynobuffers.FieldTypeBool, false)
	s.AddField("oman_default_setting", dynobuffers.FieldTypeBool, false)
	s.AddField("id_rent_periods", dynobuffers.FieldTypeInt64, false)
	s.AddField("delay_separate_mins", dynobuffers.FieldTypeInt32, false)
	s.AddField("id_ksc", dynobuffers.FieldTypeInt64, false)
	s.AddField("ml_pc_text", dynobuffers.FieldTypeString, false)   // blob
	s.AddField("ml_rm_text", dynobuffers.FieldTypeString, false)   // blob
	s.AddField("ml_oman_text", dynobuffers.FieldTypeString, false) // blob
	s.AddField("pos_article_type", dynobuffers.FieldTypeBool, false)
	s.AddField("single_free_option", dynobuffers.FieldTypeBool, false)
	s.AddField("ks_single_item", dynobuffers.FieldTypeBool, false)
	s.AddField("allergen", dynobuffers.FieldTypeBool, false)
	s.AddField("auto_resetcourse", dynobuffers.FieldTypeBool, false)
	s.AddField("block_transfer", dynobuffers.FieldTypeBool, false)
	s.AddField("id_size_modifier", dynobuffers.FieldTypeInt64, false)
	return s
}

func getSimpleScheme() *dynobuffers.Scheme {
	s, _ := dynobuffers.YamlToScheme(`
name: string
price: float32
quantity: int32
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
		panic(err)
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
		"field7": "35\"-H'1]};..1[!5Q7$$( Fk5C5!?.\"1>)/$%Tw\"E}$Mq\"\"r2Ia?Ni$\",<Xi#9\"5!t\"Be:U+#/f+-~51.,N{8\\{:I%5Fc*Hi1=,%",
		"field6": "<M#)*b'U!4L?;>x#.l766,^a*@;9Qa%*`+;n(C7$\\q9=`3&|9I5$/f,S5(/f7T38]31Wg;:,(&:)(r%#|'=f3Sw#M1/G+0V7.$n>",
		"field9": "&Co3^g.+ ,.\"-Mc;,d!/z\"(48[u+_!=b\"Fu4#8-/z0Wg&W-%.r3+l'/r3/4'^3%(()Z%)2.%\"|#Bw6Y)5!*/1`'V9!J#93*9Ea7",
		"field8": ";5$<Ce4O%1:<.=:,8~:-|++()Ie65b4-t6Mw#2b\"6r9M5(\"h756,\\y8^k>\"~<M)(V1$Ig:. 5;44Rg6@50 $890<.d(+*8>t*Z+-",
		"field5": "$Q3$22;=t:@k#U),5z%Yy'Ke(5r:S%?G!<Hg1Ii2Fi26p*E?8<j2\"`8W-126.D%)&v1!v:]w , $Bq/04 Wk%+$5Ia,$$)3x4^#4",
		"field4": "9Hq?B{\"- <Qa4Y-(Ku$<60#d##p*?6&+x/..&5l:0v'T3-Cw:., '(( :4Zw9; 8Um 80=<>7L%9+l6R?1(z??x;.4$Bu2W5%Mw7",
	}

	view["values"] = values

	data := map[string]interface{}{}

	views := make([]interface{}, 1)
	views[0] = view
	data["viewMods"] = views

	return data
}
