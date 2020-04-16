/*
 * Copyright (c) 2019-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package benchmarks

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/untillpro/dynobuffers"
)

var pbillYaml string = `
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

func BenchmarkPBillSet(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	if err != nil {
		b.Fatal(err)
	}

	pb := dynobuffers.NewBuffer(s)

	fillBuffer(pb)

	jsonBytes := []byte(pb.ToJSON())
	pb = dynobuffers.NewBuffer(s)
	dest := map[string]interface{}{}
	err = json.Unmarshal(jsonBytes, &dest)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = pb.ApplyMap(dest)
		if err != nil {
			b.Fatal(err)
		}
		_, _ = pb.ToBytes()
	}
}

func BenchmarkPBillAppend(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	if err != nil {
		b.Fatal(err)
	}

	pb := dynobuffers.NewBuffer(s)

	fillBuffer(pb)

	jsonBytes := []byte(pb.ToJSON())
	pb = dynobuffers.NewBuffer(s)
	dest := map[string]interface{}{}
	err = json.Unmarshal(jsonBytes, &dest)
	if err != nil {
		b.Fatal(err)
	}
	bytes, err := pb.ApplyJSONAndToBytes(jsonBytes)
	pb = dynobuffers.ReadBuffer(bytes, s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = pb.ApplyMap(dest)
		if err != nil {
			b.Fatal(err)
		}
		_, _ = pb.ToBytes()
		pb = dynobuffers.ReadBuffer(bytes, s)
	}
}

func BenchmarkPBillItemReadByIndex(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	if err != nil {
		b.Fatal(err)
	}

	pb := dynobuffers.NewBuffer(s)

	fillBuffer(pb)
	bytes, err := pb.ToBytes()
	if err != nil {
		b.Fatal(err)
	}

	pb = dynobuffers.ReadBuffer(bytes, s)
	pbillItem := pb.GetByIndex("pbill_item", 0).(*dynobuffers.Buffer)
	pbillItems := []*dynobuffers.Buffer{}
	for i := 0; i < 9; i++ {
		pbillItems = append(pbillItems, pbillItem)
	}
	pb.Append("pbill_item", pbillItems)

	bytes, err = pb.ToBytes()
	if err != nil {
		b.Fatal(err)
	}

	pb = dynobuffers.ReadBuffer(bytes, s)
	assert.Equal(b, 10, pb.Get("pbill_item").(*dynobuffers.ObjectArray).Len)

	sum := float32(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pb = dynobuffers.ReadBuffer(bytes, s)
		for i := 0; i < 10; i++ {
			pbillItem := pb.GetByIndex("pbill_item", i).(*dynobuffers.Buffer)
			s, _ := pbillItem.GetFloat("price")
			sum += s
		}
	}
	_ = sum
}

func BenchmarkPBillItemReadIter(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	if err != nil {
		b.Fatal(err)
	}

	pb := dynobuffers.NewBuffer(s)

	fillBuffer(pb)
	bytes, err := pb.ToBytes()
	if err != nil {
		b.Fatal(err)
	}

	pb = dynobuffers.ReadBuffer(bytes, s)
	pbillItem := pb.GetByIndex("pbill_item", 0).(*dynobuffers.Buffer)
	pbillItems := []*dynobuffers.Buffer{}
	for i := 0; i < 9; i++ {
		pbillItems = append(pbillItems, pbillItem)
	}
	pb.Append("pbill_item", pbillItems)

	bytes, err = pb.ToBytes()
	if err != nil {
		b.Fatal(err)
	}

	pb = dynobuffers.ReadBuffer(bytes, s)
	assert.Equal(b, 10, pb.Get("pbill_item").(*dynobuffers.ObjectArray).Len)

	sum := float32(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pb = dynobuffers.ReadBuffer(bytes, s)
		pbillItems := pb.Get("pbill_item").(*dynobuffers.ObjectArray)
		// pbillItems.GetDoubles()
		for pbillItems.Next() {
			s, _ := pbillItems.Buffer.GetFloat("price")
			sum += s
		}
	}
	_ = sum
}

func BenchmarkPBillItemReadNoAlloc(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	if err != nil {
		b.Fatal(err)
	}

	pb := dynobuffers.NewBuffer(s)

	fillBuffer(pb)
	bytes, err := pb.ToBytes()
	if err != nil {
		b.Fatal(err)
	}

	pb = dynobuffers.ReadBuffer(bytes, s)
	pbillItem := pb.GetByIndex("pbill_item", 0).(*dynobuffers.Buffer)
	pbillItems := []*dynobuffers.Buffer{}
	for i := 0; i < 9; i++ {
		pbillItems = append(pbillItems, pbillItem)
	}
	pb.Append("pbill_item", pbillItems)

	bytes, err = pb.ToBytes()
	if err != nil {
		b.Fatal(err)
	}

	pb = dynobuffers.ReadBuffer(bytes, s)
	assert.Equal(b, 10, pb.Get("pbill_item").(*dynobuffers.ObjectArray).Len)

	sum := float32(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pb = dynobuffers.ReadBuffer(bytes, s)
		arr := pb.Get("pbill_item").(*dynobuffers.ObjectArray)
		for arr.Next() {
			s, _ := arr.Buffer.GetFloat("price")
			sum += s
		}
	}
	_ = sum
}

func BenchmarkPbillJson(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	if err != nil {
		b.Fatal(err)
	}

	pb := dynobuffers.NewBuffer(s)

	fillBuffer(pb)

	jsonBytes := []byte(pb.ToJSON())
	dest := map[string]interface{}{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = json.Unmarshal(jsonBytes, &dest)
		if err != nil {
			b.Fatal(err)
		}
		_, _ = json.Marshal(dest)
		dest = map[string]interface{}{}
	}
}

func fillBuffer(b *dynobuffers.Buffer) {
	for i, f := range b.Scheme.Fields {
		var val interface{}
		switch f.Ft {
		case dynobuffers.FieldTypeBool:
			val = true
		case dynobuffers.FieldTypeString:
			val = "str" + strconv.Itoa(i)
		case dynobuffers.FieldTypeObject:
			nested := dynobuffers.NewBuffer(f.FieldScheme)
			fillBuffer(nested)
			if f.IsArray {
				val = []*dynobuffers.Buffer{nested}
			} else {
				val = nested
			}
		case dynobuffers.FieldTypeByte:
			if f.IsArray {
				val = []byte{byte(i)}
			} else {
				val = i
			}
		default:
			val = i
		}
		b.Set(f.Name, val)
	}
}
