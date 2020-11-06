/*
 * Copyright (c) 2019-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

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
Id: long
Id_bill: long
Id_untill_users: long
number: int
failurednumber: int
suffix: string
pdatetime: string
id_sales_area: long
pcname: string
service_charge: double
real_datetime: string
tips: float
id_clients: long
pbill_index: int
split_parts: double
id_real_untill_user: long
reopen_type: byte
external_id: string
covers: int
hht: string
super_dx: long
c_tips: float
id_currency: long
bill:
  Id: long
  Tableno: int
  Id_untill_users: long
  Table_part: string
  id_courses: long
  id_clients: long
  name: string
  Proforma: byte
  modified: string
  open_datetime: string
  close_datetime: string
  number: int
  failurednumber: int
  suffix: string
  pbill_number: int
  pbill_failurednumber: int
  pbill_suffix: string
  hc_foliosequence: int
  hc_folionumber: string
  tip: float
  qty_persons: int
  isdirty: byte
  reservationid: string
  id_alter_user: long
  service_charge: double
  number_of_covers: int
  id_user_proforma: long
  bill_type: byte
  locker: int
  id_time_article: long
  timer_start: string
  timer_stop: string
  isactive: byte
  table_name: string
  group_vat_level: byte
  comments: string
  id_cardprice: long
  discount: double
  discount_value: float
  id_discount_reasons: long
  hc_roomnumber: string
  ignore_auto_sc: byte
  extra_fields..: byte
  id_bo_service_charge: long
  free_comments: string
  id_t2o_groups: long
  service_tax: float
  sc_plan..: byte
  client_phone: string
  age: string
  description..: byte
  sdescription: string
  vars..: byte
  take_away: int
  fiscal_number: int
  fiscal_failurednumber: int
  fiscal_suffix: string
  id_order_type: long
  not_paid: float
  total: float
bill_reprints..:
  Id: long
  Id_pbill: long
  datetime: string
  id_untill_users: long
bill_split_rest..:
  Id: long
  id_bill: long
  is_active: byte
  split_data..: byte
  datetime: string
  id_pbill: long
complete_meal..:
  Id: long
  id_pbill: long
  id_untill_users: long
  cm_datetime: string
  cm_pcname: string
  cm_quantity: int
  cm_price: float
neg_pbill..:
  Id: long
  Id_pbill: long
  R_type: byte
  pdatetime: string
  id_pbill_close: long
open_discounts..:
  Id: long
  Id_pbill: long
  amount: float
  vat_percent: float
  vat: double
pbill_balance..:
  Id: long
  id_pbill: long
  id_clients: long
  id_accounts: long
  pre_balance: float
  pre_balance_common: float
pbill_cleancash_info..:
  Id: long
  id_pbill: long
  clean_cash_trnumber: long
  clean_cash_signature: string
  manufacturing_code: string
  control_timestamp: string
  ticket_counter: string
  vsc_id: string
  plu_data_hash: string
  id_bill: long
  ccdatetime: string
  kind: byte
  production_nr: string
  extra1: string
  extra2: string
pbill_email..:
  Id: long
  id_pbill: long
  pemail_email: string
  pemail_body..: byte
  pemail_dt: string
  pemail_state: byte
pbill_item..:
  Id: long
  Id_pbill: long
  Id_untill_users: long
  Tableno: int
  Rowbeg: byte
  Kind: byte
  Quantity: int
  id_articles: long
  Price: float
  text: string
  id_prices: long
  with_without: int
  id_courses: long
  pdatetime: string
  purchase_price: float
  vat_sign: string
  vat: double
  id_menu: long
  menu_quantity: int
  original_price: float
  id_discount_reasons: long
  discount_type: int
  chair_number: int
  separated: byte
  negative: byte
  vat_percent: float
  chair_name: string
  id_option_article: long
  start_delay_minutes: int
  pua: byte
  no_price: byte
  orig_pbill_id: long
  parts_id: long
  full_paid: byte
  weight: float
  splitted_coef: double
  split_id: long
  free_comments: string
  id_smartcards: long
  has_allergens: byte
  id_article_options: long
  order_item_allergens..:
    Id: long
    id_order_item: long
    id_menu_item: long
    id_allergens: long
    text: string
    id_pbill_item: long
  order_item_extra..:
    Id: long
    id_order_item: long
    id_menu_item: long
    id_pbill_item: long
    id_article_barcodes: long
  order_item_sizes..:
    Id: long
    id_order_item: long
    id_menu_item: long
    id_pbill_item: long
    id_articles: long
    price: float
    original_price: float
    vat: double
    id_size_modifier_item: long
  pbill_hash..:
    Id: long
    Id_pbill_item: long
    Id_untill_users: long
  pbill_item_bookp..:
    Id: long
    id_pbill_item: long
    id_bookkeeping_turnover: long
    id_bookkeeping_vat: long
pbill_item_refund..:
  Id: long
  Id_pbill: long
  Id_untill_users: long
  Tableno: int
  Rowbeg: byte
  Kind: byte
  Quantity: int
  id_articles: long
  id_department: long
  id_food_group: long
  id_category: long
  Price: float
  text: string
  id_prices: long
  with_without: int
  id_courses: long
  pdatetime: string
  purchase_price: float
  vat_sign: string
  vat: double
  id_menu: long
  menu_quantity: int
  ref_type: byte
pbill_payments..:
  Id: long
  Id_pbill: long
  Id_payments: long
  Price: float
  customer_amount: float
  c_price: float
  c_customer_amount: float
  vat: float
  id_driver_discount: long
  accounts_payments..:
    Id: long
    datetime: string
    Id_clients: long
    Id_untill_users: long
    tableno: int
    tablepart: string
    number: int
    suffix: string
    name: string
    tr_datetime: string
    bill_number: string
    parts_count: byte
    part_index: byte
    discount: double
    id_sales_area: long
    total: float
    account_type: byte
    payer_name: string
    id_pbill_payments: long
  account_payment_data..:
    Id: long
    id_pbill_payments: long
    account_type: byte
  payments_hash..:
    Id: long
    Id_payments: long
    Id_untill_users: long
  pbill_card_payments_info..:
    Id: long
    Id_pbill_payments: long
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
    loyaltyamount: long
    controlno: string
    chargetype: int
    cardname: string
    cardnoentertype: int
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
    cashback_amount: long
    field11: string
    field12: string
    field13: string
    field14: string
    field15: string
  pbill_payer..:
    Id: long
    id_pbill_payments: long
    payer_name: string
  pbill_payments_bookp..:
    Id: long
    id_pbill_payments: long
    id_bookkeeping: long
  pbill_zapper_payments_info..:
    Id: long
    Id_pbill_payments: long
    Id_zapper_proformas: long
    zapper_id: string
    misc_data..: byte
  psp_tips..:
    Id: long
    Id_pbill: long
    amount: float
    id_pbill_payments: long
  smartcards_turnover..:
    Id: long
    datetime: string
    Id_smartcards: long
    Id_untill_users: long
    kind: byte
    amount: float
    p_amount: float
    id_pbill_payments: long
    id_payments: long
    is_initial_value: byte
    card_data..: byte
    deposit_number: int
    deposit_suffix: string
    pcname: string
    new_balance: float
  voucher_payments..:
    Id: long
    id_vouchers: long
    id_pbill_payments: long
    used_dt: string
  voucher_pbill_payments..:
    Id: long
    id_voucher_pbill: long
    price: float
    id_vouchers: long
    id_payments: long
    id_pbill_payments: long
pbill_return..:
  Id: long
  id_pbill: long
  ref_type: byte
  id_void_reason: long
  void_text: string
psp_tips..:
  Id: long
  Id_pbill: long
  amount: float
  id_pbill_payments: long
sold_articles..:
  Id: long
  id_pbill: long
  id_order_item: long
  quantity: int
  parts_id: long
  sa_coef: double
  split_id: long
sold_articles_wm..:
  Id: long
  id_order_item: long
  id_pbill: long
  quantity: int
  sa_coef: double
stock_doc..:
  Id: long
  id_orders: long
  id_pbill: long
  id_stock_invoice: long
  id_stock_adjustment: long
  id_stock_cost_correction: long
stock_pbill_queue..:
  Id: long
  Id_pbill: long
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
	s.AddField("id", dynobuffers.FieldTypeLong, false)
	s.AddField("article_number", dynobuffers.FieldTypeInt, false)
	s.AddField("name", dynobuffers.FieldTypeString, false)
	s.AddField("internal_name", dynobuffers.FieldTypeString, false)
	s.AddField("article_manual", dynobuffers.FieldTypeBool, false)
	s.AddField("article_hash", dynobuffers.FieldTypeBool, false)
	s.AddField("id_courses", dynobuffers.FieldTypeLong, false)
	s.AddField("id_departament", dynobuffers.FieldTypeLong, false)
	s.AddField("pc_bitmap", dynobuffers.FieldTypeString, false) // blob
	s.AddField("pc_color", dynobuffers.FieldTypeInt, false)
	s.AddField("pc_text", dynobuffers.FieldTypeString, false)
	s.AddField("pc_font_name", dynobuffers.FieldTypeString, false)
	s.AddField("pc_font_size", dynobuffers.FieldTypeInt, false)
	s.AddField("pc_font_attr", dynobuffers.FieldTypeInt, false)
	s.AddField("pc_font_color", dynobuffers.FieldTypeInt, false)
	s.AddField("rm_text", dynobuffers.FieldTypeString, false)
	s.AddField("rm_font_size", dynobuffers.FieldTypeInt, false)
	s.AddField("id_packing", dynobuffers.FieldTypeLong, false)
	s.AddField("id_commission", dynobuffers.FieldTypeLong, false)
	s.AddField("id_promotions", dynobuffers.FieldTypeLong, false)
	s.AddField("savepoints", dynobuffers.FieldTypeInt, false)
	s.AddField("quantity", dynobuffers.FieldTypeInt, false)
	s.AddField("hideonhold", dynobuffers.FieldTypeBool, false)
	s.AddField("barcode", dynobuffers.FieldTypeString, false)
	s.AddField("time_active", dynobuffers.FieldTypeBool, false)
	s.AddField("aftermin", dynobuffers.FieldTypeInt, false)
	s.AddField("periodmin", dynobuffers.FieldTypeInt, false)
	s.AddField("roundmin", dynobuffers.FieldTypeInt, false)
	s.AddField("id_currency", dynobuffers.FieldTypeLong, false)
	s.AddField("control_active", dynobuffers.FieldTypeBool, false)
	s.AddField("control_time", dynobuffers.FieldTypeInt, false)
	s.AddField("plu_number_vanduijnen", dynobuffers.FieldTypeInt, false)
	s.AddField("sequence", dynobuffers.FieldTypeInt, false)
	s.AddField("rm_sequence", dynobuffers.FieldTypeInt, false)
	s.AddField("purchase_price", dynobuffers.FieldTypeFloat, false)
	s.AddField("id_vd_group", dynobuffers.FieldTypeLong, false)
	s.AddField("menu", dynobuffers.FieldTypeBool, false)
	s.AddField("sensitive", dynobuffers.FieldTypeBool, false)
	s.AddField("sensitive_option", dynobuffers.FieldTypeBool, false)
	s.AddField("daily_stock", dynobuffers.FieldTypeInt, false)
	s.AddField("info", dynobuffers.FieldTypeString, false)
	s.AddField("warning_level", dynobuffers.FieldTypeInt, false)
	s.AddField("free_after_pay", dynobuffers.FieldTypeInt, false)
	s.AddField("id_food_group", dynobuffers.FieldTypeLong, false)
	s.AddField("article_type", dynobuffers.FieldTypeByte, false)
	s.AddField("id_inventory_item", dynobuffers.FieldTypeLong, false)
	s.AddField("id_recipe", dynobuffers.FieldTypeLong, false)
	s.AddField("id_unity_sales", dynobuffers.FieldTypeLong, false)
	s.AddField("can_savepoints", dynobuffers.FieldTypeBool, false)
	s.AddField("show_in_kitchen_screen", dynobuffers.FieldTypeBool, false)
	s.AddField("decrease_savepoints", dynobuffers.FieldTypeInt, false)
	s.AddField("hht_color", dynobuffers.FieldTypeInt, false)
	s.AddField("hht_font_name", dynobuffers.FieldTypeString, false)
	s.AddField("hht_font_size", dynobuffers.FieldTypeInt, false)
	s.AddField("hht_font_attr", dynobuffers.FieldTypeInt, false)
	s.AddField("hht_font_color", dynobuffers.FieldTypeInt, false)
	s.AddField("tip", dynobuffers.FieldTypeBool, false)
	s.AddField("id_beco_group", dynobuffers.FieldTypeLong, false)
	s.AddField("id_beco_location", dynobuffers.FieldTypeLong, false)
	s.AddField("bc_standard_dosage", dynobuffers.FieldTypeInt, false)
	s.AddField("bc_alternative_dosage", dynobuffers.FieldTypeInt, false)
	s.AddField("bc_disablebalance", dynobuffers.FieldTypeBool, false)
	s.AddField("bc_use_locations", dynobuffers.FieldTypeBool, false)
	s.AddField("time_rate", dynobuffers.FieldTypeFloat, false)
	s.AddField("id_free_option", dynobuffers.FieldTypeLong, false)
	s.AddField("party_article", dynobuffers.FieldTypeBool, false)
	s.AddField("id_pua_groups", dynobuffers.FieldTypeLong, false)
	s.AddField("promo", dynobuffers.FieldTypeBool, false)
	s.AddField("one_hand_limit", dynobuffers.FieldTypeInt, false)
	s.AddField("consolidate_quantity", dynobuffers.FieldTypeInt, false)
	s.AddField("consolidate_alias_name", dynobuffers.FieldTypeString, false)
	s.AddField("hq_id", dynobuffers.FieldTypeString, false)
	s.AddField("is_active", dynobuffers.FieldTypeBool, false)
	s.AddField("is_active_modified", dynobuffers.FieldTypeInt, false) // timestamp
	s.AddField("is_active_modifier", dynobuffers.FieldTypeString, false)
	s.AddField("rent_price_type", dynobuffers.FieldTypeBool, false)
	s.AddField("id_rental_group", dynobuffers.FieldTypeLong, false)
	s.AddField("condition_check_in_order", dynobuffers.FieldTypeBool, false)
	s.AddField("weight_required", dynobuffers.FieldTypeBool, false)
	s.AddField("daily_numeric_1", dynobuffers.FieldTypeFloat, false)
	s.AddField("daily_numeric_2", dynobuffers.FieldTypeFloat, false)
	s.AddField("prep_min", dynobuffers.FieldTypeInt, false)
	s.AddField("id_article_ksp", dynobuffers.FieldTypeLong, false)
	s.AddField("warn_min", dynobuffers.FieldTypeInt, false)
	s.AddField("empty_article", dynobuffers.FieldTypeBool, false)
	s.AddField("bc_debitcredit", dynobuffers.FieldTypeBool, false)
	s.AddField("prep_sec", dynobuffers.FieldTypeInt, false)
	s.AddField("id_suppliers", dynobuffers.FieldTypeLong, false)
	s.AddField("main_price", dynobuffers.FieldTypeBool, false)
	s.AddField("oman_text", dynobuffers.FieldTypeString, false)
	s.AddField("id_age_groups", dynobuffers.FieldTypeLong, false)
	s.AddField("surcharge", dynobuffers.FieldTypeBool, false)
	s.AddField("info_data", dynobuffers.FieldTypeString, false) //blob
	s.AddField("pos_disabled", dynobuffers.FieldTypeBool, false)
	s.AddField("ml_name", dynobuffers.FieldTypeString, false)    // blob
	s.AddField("ml_ks_name", dynobuffers.FieldTypeString, false) // blob
	s.AddField("alt_articles", dynobuffers.FieldTypeLong, false)
	s.AddField("alt_alias", dynobuffers.FieldTypeString, false)
	s.AddField("need_prep", dynobuffers.FieldTypeBool, false)
	s.AddField("auto_onhold", dynobuffers.FieldTypeBool, false)
	s.AddField("id_ks_wf", dynobuffers.FieldTypeLong, false)
	s.AddField("ks_wf_type", dynobuffers.FieldTypeInt, false)
	s.AddField("ask_course", dynobuffers.FieldTypeBool, false)
	s.AddField("popup_info", dynobuffers.FieldTypeString, false)
	s.AddField("allow_order_items", dynobuffers.FieldTypeBool, false)
	s.AddField("must_combined", dynobuffers.FieldTypeBool, false)
	s.AddField("block_discount", dynobuffers.FieldTypeBool, false)
	s.AddField("has_default_options", dynobuffers.FieldTypeBool, false)
	s.AddField("hht_default_setting", dynobuffers.FieldTypeBool, false)
	s.AddField("oman_default_setting", dynobuffers.FieldTypeBool, false)
	s.AddField("id_rent_periods", dynobuffers.FieldTypeLong, false)
	s.AddField("delay_separate_mins", dynobuffers.FieldTypeInt, false)
	s.AddField("id_ksc", dynobuffers.FieldTypeLong, false)
	s.AddField("ml_pc_text", dynobuffers.FieldTypeString, false)   // blob
	s.AddField("ml_rm_text", dynobuffers.FieldTypeString, false)   // blob
	s.AddField("ml_oman_text", dynobuffers.FieldTypeString, false) // blob
	s.AddField("pos_article_type", dynobuffers.FieldTypeBool, false)
	s.AddField("single_free_option", dynobuffers.FieldTypeBool, false)
	s.AddField("ks_single_item", dynobuffers.FieldTypeBool, false)
	s.AddField("allergen", dynobuffers.FieldTypeBool, false)
	s.AddField("auto_resetcourse", dynobuffers.FieldTypeBool, false)
	s.AddField("block_transfer", dynobuffers.FieldTypeBool, false)
	s.AddField("id_size_modifier", dynobuffers.FieldTypeLong, false)
	return s
}

func getSimpleScheme() *dynobuffers.Scheme {
	s, _ := dynobuffers.YamlToScheme(`
name: string
price: float
quantity: int
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
