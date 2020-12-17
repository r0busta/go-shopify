package goshopify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func productTests(t *testing.T, product Product) {
	// Check that ID is assigned to the returned product
	var expectedInt int64 = 1071559748
	if product.ID != expectedInt {
		t.Errorf("Product.ID returned %+v, expected %+v", product.ID, expectedInt)
	}
}

func TestProductList(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://fooshop.myshopify.com/%s/products.json", client.pathPrefix),
		httpmock.NewStringResponder(200, `{"products": [{"id":1},{"id":2}]}`))

	products, _, err := client.Product.List(nil)
	if err != nil {
		t.Errorf("Product.List returned error: %v", err)
	}

	expected := []*Product{{ID: 1}, {ID: 2}}
	if !reflect.DeepEqual(products, expected) {
		t.Errorf("Product.List returned %+v, expected %+v", products, expected)
	}
}

func TestProductListFilterByIds(t *testing.T) {
	setup()
	defer teardown()

	params := map[string]string{"ids": "1,2,3"}
	httpmock.RegisterResponderWithQuery(
		"GET",
		fmt.Sprintf("https://fooshop.myshopify.com/%s/products.json", client.pathPrefix),
		params,
		httpmock.NewStringResponder(200, `{"products": [{"id":1},{"id":2},{"id":3}]}`))

	listOptions := ListOptions{IDs: []int64{1, 2, 3}}

	products, _, err := client.Product.List(listOptions)
	if err != nil {
		t.Errorf("Product.List returned error: %v", err)
	}

	expected := []*Product{{ID: 1}, {ID: 2}, {ID: 3}}
	if !reflect.DeepEqual(products, expected) {
		t.Errorf("Product.List returned %+v, expected %+v", products, expected)
	}
}

func TestProductListError(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://fooshop.myshopify.com/%s/products.json", client.pathPrefix),
		httpmock.NewStringResponder(500, ""))

	expectedErrMessage := "Unknown Error"

	products, err := client.Product.List(nil)
	if products != nil {
		t.Errorf("Product.List returned products, expected nil: %v", err)
	}

	if err == nil || err.Error() != expectedErrMessage {
		t.Errorf("Product.List err returned %+v, expected %+v", err, expectedErrMessage)
	}
}

func TestProductListWithPagination(t *testing.T) {
	setup()
	defer teardown()

	listURL := fmt.Sprintf("https://fooshop.myshopify.com/%s/products.json", client.pathPrefix)

	// The strconv.Atoi error changed in go 1.8, 1.7 is still being tested/supported.
	limitConversionErrorMessage := `strconv.Atoi: parsing "invalid": invalid syntax`
	if runtime.Version()[2:5] == "1.7" {
		limitConversionErrorMessage = `strconv.ParseInt: parsing "invalid": invalid syntax`
	}

	cases := []struct {
		body               string
		linkHeader         string
		expectedProducts   []Product
		expectedPagination *Pagination
		expectedErr        error
	}{
		// Expect empty pagination when there is no link header
		{
			`{"products": [{"id":1},{"id":2}]}`,
			"",
			[]Product{{ID: 1}, {ID: 2}},
			new(Pagination),
			nil,
		},
		// Invalid link header responses
		{
			"{}",
			"invalid link",
			[]Product(nil),
			nil,
			ResponseDecodingError{Message: "could not extract pagination link header"},
		},
		{
			"{}",
			`<:invalid.url>; rel="next"`,
			[]Product(nil),
			nil,
			ResponseDecodingError{Message: "pagination does not contain a valid URL"},
		},
		{
			"{}",
			`<http://valid.url?%invalid_query>; rel="next"`,
			[]Product(nil),
			nil,
			errors.New(`invalid URL escape "%in"`),
		},
		{
			"{}",
			`<http://valid.url>; rel="next"`,
			[]Product(nil),
			nil,
			ResponseDecodingError{Message: "page_info is missing"},
		},
		{
			"{}",
			`<http://valid.url?page_info=foo&limit=invalid>; rel="next"`,
			[]Product(nil),
			nil,
			errors.New(limitConversionErrorMessage),
		},
		// Valid link header responses
		{
			`{"products": [{"id":1}]}`,
			`<http://valid.url?page_info=foo&limit=2>; rel="next"`,
			[]Product{{ID: 1}},
			&Pagination{
				NextPageOptions: &ListOptions{PageInfo: "foo", Limit: 2},
			},
			nil,
		},
		{
			`{"products": [{"id":2}]}`,
			`<http://valid.url?page_info=foo>; rel="next", <http://valid.url?page_info=bar>; rel="previous"`,
			[]Product{{ID: 2}},
			&Pagination{
				NextPageOptions:     &ListOptions{PageInfo: "foo"},
				PreviousPageOptions: &ListOptions{PageInfo: "bar"},
			},
			nil,
		},
	}
	for i, c := range cases {
		response := &http.Response{
			StatusCode: 200,
			Body:       httpmock.NewRespBodyFromString(c.body),
			Header: http.Header{
				"Link": {c.linkHeader},
			},
		}

		httpmock.RegisterResponder("GET", listURL, httpmock.ResponderFromResponse(response))

		products, pagination, err := client.Product.ListWithPagination(nil)
		if !reflect.DeepEqual(products, c.expectedProducts) {
			t.Errorf("test %d Product.ListWithPagination products returned %+v, expected %+v", i, products, c.expectedProducts)
		}

		if !reflect.DeepEqual(pagination, c.expectedPagination) {
			t.Errorf(
				"test %d Product.ListWithPagination pagination returned %+v, expected %+v",
				i,
				pagination,
				c.expectedPagination,
			)
		}

		if (c.expectedErr != nil || err != nil) && err.Error() != c.expectedErr.Error() {
			t.Errorf(
				"test %d Product.ListWithPagination err returned %+v, expected %+v",
				i,
				err,
				c.expectedErr,
			)
		}
	}
}

func TestProductCount(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://fooshop.myshopify.com/%s/products/count.json", client.pathPrefix),
		httpmock.NewStringResponder(200, `{"count": 3}`))

	params := map[string]string{"created_at_min": "2016-01-01T00:00:00Z"}
	httpmock.RegisterResponderWithQuery(
		"GET",
		fmt.Sprintf("https://fooshop.myshopify.com/%s/products/count.json", client.pathPrefix),
		params,
		httpmock.NewStringResponder(200, `{"count": 2}`))

	cnt, err := client.Product.Count(nil)
	if err != nil {
		t.Errorf("Product.Count returned error: %v", err)
	}

	expected := 3
	if cnt != expected {
		t.Errorf("Product.Count returned %d, expected %d", cnt, expected)
	}

	date := time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC)
	cnt, err = client.Product.Count(CountOptions{CreatedAtMin: date})
	if err != nil {
		t.Errorf("Product.Count returned error: %v", err)
	}

	expected = 2
	if cnt != expected {
		t.Errorf("Product.Count returned %d, expected %d", cnt, expected)
	}
}

func TestProductGet(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://fooshop.myshopify.com/%s/products/1.json", client.pathPrefix),
		httpmock.NewStringResponder(200, `{"product": {"id":1}}`))

	product, err := client.Product.Get(1, nil)
	if err != nil {
		t.Errorf("Product.Get returned error: %v", err)
	}

	expected := &Product{ID: 1}
	if !reflect.DeepEqual(product, expected) {
		t.Errorf("Product.Get returned %+v, expected %+v", product, expected)
	}
}

func TestProductCreate(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("POST", fmt.Sprintf("https://fooshop.myshopify.com/%s/products.json", client.pathPrefix),
		httpmock.NewBytesResponder(200, loadFixture("product.json")))

	product := Product{
		Title:       "Burton Custom Freestyle 151",
		BodyHTML:    "<strong>Good snowboard!<\\/strong>",
		Vendor:      "Burton",
		ProductType: "Snowboard",
	}

	returnedProduct, err := client.Product.Create(product)
	if err != nil {
		t.Errorf("Product.Create returned error: %v", err)
	}

	productTests(t, *returnedProduct)
}

func TestProductUpdate(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("PUT", fmt.Sprintf("https://fooshop.myshopify.com/%s/products/1.json", client.pathPrefix),
		httpmock.NewBytesResponder(200, loadFixture("product.json")))

	product := Product{
		ID:          1,
		ProductType: "Skateboard",
	}

	returnedProduct, err := client.Product.Update(product)
	if err != nil {
		t.Errorf("Product.Update returned error: %v", err)
	}

	productTests(t, *returnedProduct)
}

func TestProductDelete(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("DELETE", fmt.Sprintf("https://fooshop.myshopify.com/%s/products/1.json", client.pathPrefix),
		httpmock.NewStringResponder(200, "{}"))

	err := client.Product.Delete(1)
	if err != nil {
		t.Errorf("Product.Delete returned error: %v", err)
	}
}

func TestProductListMetafields(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://fooshop.myshopify.com/%s/products/1/metafields.json", client.pathPrefix),
		httpmock.NewStringResponder(200, `{"metafields": [{"id":1},{"id":2}]}`))

	metafields, err := client.Product.ListMetafields(1, nil)
	if err != nil {
		t.Errorf("Product.ListMetafields() returned error: %v", err)
	}

	expected := []Metafield{{ID: 1}, {ID: 2}}
	if !reflect.DeepEqual(metafields, expected) {
		t.Errorf("Product.ListMetafields() returned %+v, expected %+v", metafields, expected)
	}
}

func TestProductCountMetafields(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://fooshop.myshopify.com/%s/products/1/metafields/count.json", client.pathPrefix),
		httpmock.NewStringResponder(200, `{"count": 3}`))

	params := map[string]string{"created_at_min": "2016-01-01T00:00:00Z"}
	httpmock.RegisterResponderWithQuery(
		"GET",
		fmt.Sprintf("https://fooshop.myshopify.com/%s/products/1/metafields/count.json", client.pathPrefix),
		params,
		httpmock.NewStringResponder(200, `{"count": 2}`))

	cnt, err := client.Product.CountMetafields(1, nil)
	if err != nil {
		t.Errorf("Product.CountMetafields() returned error: %v", err)
	}

	expected := 3
	if cnt != expected {
		t.Errorf("Product.CountMetafields() returned %d, expected %d", cnt, expected)
	}

	date := time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC)
	cnt, err = client.Product.CountMetafields(1, CountOptions{CreatedAtMin: date})
	if err != nil {
		t.Errorf("Product.CountMetafields() returned error: %v", err)
	}

	expected = 2
	if cnt != expected {
		t.Errorf("Product.CountMetafields() returned %d, expected %d", cnt, expected)
	}
}

func TestProductGetMetafield(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://fooshop.myshopify.com/%s/products/1/metafields/2.json", client.pathPrefix),
		httpmock.NewStringResponder(200, `{"metafield": {"id":2}}`))

	metafield, err := client.Product.GetMetafield(1, 2, nil)
	if err != nil {
		t.Errorf("Product.GetMetafield() returned error: %v", err)
	}

	expected := &Metafield{ID: 2}
	if !reflect.DeepEqual(metafield, expected) {
		t.Errorf("Product.GetMetafield() returned %+v, expected %+v", metafield, expected)
	}
}

func TestProductCreateMetafield(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("POST", fmt.Sprintf("https://fooshop.myshopify.com/%s/products/1/metafields.json", client.pathPrefix),
		httpmock.NewBytesResponder(200, loadFixture("metafield.json")))

	metafield := Metafield{
		Key:       "app_key",
		Value:     "app_value",
		ValueType: "string",
		Namespace: "affiliates",
	}

	returnedMetafield, err := client.Product.CreateMetafield(1, metafield)
	if err != nil {
		t.Errorf("Product.CreateMetafield() returned error: %v", err)
	}

	MetafieldTests(t, *returnedMetafield)
}

func TestProductUpdateMetafield(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("PUT", fmt.Sprintf("https://fooshop.myshopify.com/%s/products/1/metafields/2.json", client.pathPrefix),
		httpmock.NewBytesResponder(200, loadFixture("metafield.json")))

	metafield := Metafield{
		ID:        2,
		Key:       "app_key",
		Value:     "app_value",
		ValueType: "string",
		Namespace: "affiliates",
	}

	returnedMetafield, err := client.Product.UpdateMetafield(1, metafield)
	if err != nil {
		t.Errorf("Product.UpdateMetafield() returned error: %v", err)
	}

	MetafieldTests(t, *returnedMetafield)
}

func TestProductDeleteMetafield(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("DELETE", fmt.Sprintf("https://fooshop.myshopify.com/%s/products/1/metafields/2.json", client.pathPrefix),
		httpmock.NewStringResponder(200, "{}"))

	err := client.Product.DeleteMetafield(1, 2)
	if err != nil {
		t.Errorf("Product.DeleteMetafield() returned error: %v", err)
	}
}

func TestEmptiableTime_MarshalJSON(t *testing.T) {
	expectedZero := time.Time{}

	expected, err := time.Parse("2006-01-02T15:04:05-07:00", "2007-12-31T19:00:00-05:00")
	if err != nil {
		t.Errorf("%s", err)
	}

	testCases := []struct {
		desc        string
		data        *EmptiableTime
		expected    string
		expectedErr bool
		equal       bool
	}{
		{"Null", nil, `null`, false, true},
		{"Empty", &EmptiableTime{}, `null`, false, true},
		{"Zero", &EmptiableTime{expectedZero}, `null`, false, true},
		{"Reference", &EmptiableTime{expected}, `"2007-12-31T19:00:00-05:00"`, false, true},
		{"Mismatch", &EmptiableTime{}, `"2020-12-31T19:00:00-05:00"`, false, false},
	}

	for _, tc := range testCases {
		out, err := json.Marshal(tc.data)
		if actualErr := err != nil; actualErr != tc.expectedErr {
			t.Errorf("%s: actualErr=%v, expectedErr=%v, err=%v", tc.desc, actualErr, tc.expectedErr, err)
			continue
		}
		actual := string(out)
		equal := actual == tc.expected
		if equal != tc.equal {
			t.Errorf("%s: actual=%#v, expected=%#v, equal=%v, expected=%v", tc.desc, actual, tc.expected, equal, tc.equal)
			continue
		}
	}
}

func TestEmptiableTime_UnmarshalJSON(t *testing.T) {
	expectedZero := time.Time{}

	expected, err := time.Parse("2006-01-02T15:04:05-07:00", "2007-12-31T19:00:00-05:00")
	if err != nil {
		t.Errorf("%s", err)
	}

	testCases := []struct {
		desc        string
		data        string
		expected    *EmptiableTime
		expectedErr bool
		equal       bool
	}{
		{"Null", `null`, nil, false, true},
		{"Empty", `""`, &EmptiableTime{expectedZero}, false, true},
		{"Zero", `"0001-01-01T00:00:00Z"`, &EmptiableTime{expectedZero}, false, true},
		{"Reference", `"2007-12-31T19:00:00-05:00"`, &EmptiableTime{expected}, false, true},
		{"Mismatch", `"2020-12-31T19:00:00-05:00"`, &EmptiableTime{}, false, false},
		{"Invalid", `"abcd"`, &EmptiableTime{expected}, true, false},
	}

	for _, tc := range testCases {
		var actual *EmptiableTime
		err := json.Unmarshal([]byte(tc.data), &actual)
		if actualErr := err != nil; actualErr != tc.expectedErr {
			t.Errorf("%s: actualErr=%v, expectedErr=%v, err=%v", tc.desc, actualErr, tc.expectedErr, err)
			continue
		}
		if actual == nil {
			if tc.expected != nil {
				t.Errorf("%s: actual=%#v, expected=%#v, equal=%v, expected=%v", tc.desc, actual, tc.expected, false, tc.equal)
				continue
			}
			continue
		}
		equal := actual.Equal(tc.expected.Time)
		if equal != tc.equal {
			t.Errorf("%s: actual=%#v, expected=%#v, equal=%v, expected=%v", tc.desc, actual, tc.expected, equal, tc.equal)
			continue
		}
	}
}

// func TestProductPublishedAt_MarshalJSON(t *testing.T) {
// 	expectedZero := &EmptiableTime{time.Time{}}

// 	publishedAt, err := time.Parse("2006-01-02T15:04:05-07:00", "2007-12-31T19:00:00-05:00")
// 	if err != nil {
// 		t.Errorf("%s", err)
// 	}

// 	expected := &EmptiableTime{publishedAt}

// 	testCases := []struct {
// 		desc        string
// 		data        *Product
// 		expected    string
// 		expectedErr bool
// 		equal       bool
// 	}{
// 		{"Null", nil, `null`, false, true},
// 		{"Unset", &Product{}, `{}`, false, true},
// 		{"Set to nil", &Product{PublishedAt: nil}, `{}`, false, true},
// 		{"Empty", &Product{PublishedAt: &EmptiableTime{}}, `{"published_at":null}`, false, true},
// 		{"Zero", &Product{PublishedAt: expectedZero}, `{"published_at":null}`, false, true},
// 		{"Reference", &Product{PublishedAt: expected}, `{"published_at":"2007-12-31T19:00:00-05:00"}`, false, true},
// 		{"Mismatch", &Product{Title: "foo"}, `{"published_at":"2020-12-31T19:00:00-05:00"}`, false, false},
// 	}

// 	for _, tc := range testCases {
// 		out, err := json.Marshal(tc.data)
// 		if actualErr := err != nil; actualErr != tc.expectedErr {
// 			t.Errorf("%s: actualErr=%v, expectedErr=%v, err=%v", tc.desc, actualErr, tc.expectedErr, err)
// 			continue
// 		}
// 		actual := string(out)
// 		equal := actual == tc.expected
// 		if equal != tc.equal {
// 			t.Errorf("%s: actual=%#v, expected=%#v, equal=%v, expected=%v", tc.desc, actual, tc.expected, equal, tc.equal)
// 			continue
// 		}
// 	}
// }

// func TestProductPublishedAt_UnmarshalJSON(t *testing.T) {
// 	publishedAt, err := time.Parse("2006-01-02T15:04:05-07:00", "2007-12-31T19:00:00-05:00")
// 	if err != nil {
// 		t.Errorf("%s", err)
// 	}

// 	expected := &EmptiableTime{publishedAt}

// 	testCases := []struct {
// 		desc        string
// 		data        string
// 		expected    *Product
// 		expectedErr bool
// 		equal       bool
// 	}{
// 		{"Null", `null`, nil, false, true},
// 		{"Set to null", `{"published_at":null}`, &Product{PublishedAt: &EmptiableTime{}}, false, true},
// 		{"Empty", `{"published_at":""}`, &Product{PublishedAt: &EmptiableTime{}}, false, true},
// 		{"Zero", `{"published_at":"0001-01-01T00:00:00Z"}`, &Product{PublishedAt: &EmptiableTime{}}, false, true},
// 		{"Reference", `{"published_at":"2007-12-31T19:00:00-05:00"}`, &Product{PublishedAt: expected}, false, true},
// 		{"Mismatch", `{"published_at":"2020-12-31T19:00:00-05:00"}`, &Product{PublishedAt: expected}, false, false},
// 		// {"Invalid", `"abcd"`, nil, true, false},
// 	}

// 	for _, tc := range testCases {
// 		var actual *Product
// 		err := json.Unmarshal([]byte(tc.data), &actual)
// 		if actualErr := err != nil; actualErr != tc.expectedErr {
// 			t.Errorf("%s: actualErr=%v, expectedErr=%v, err=%v", tc.desc, actualErr, tc.expectedErr, err)
// 			continue
// 		}
// 		if actual == nil {
// 			if tc.expected != nil {
// 				t.Errorf("%s: actual=%#v, expected=%#v, equal=%v, expected=%v", tc.desc, actual, tc.expected, false, tc.equal)
// 				continue
// 			}
// 			continue
// 		}
// 		if actual.PublishedAt == nil {
// 			if tc.expected.PublishedAt != nil {
// 				t.Errorf("%s: actual=%#v, expected=%#v, equal=%v, expected=%v", tc.desc, actual, tc.expected, false, tc.equal)
// 				continue
// 			}
// 		} else {
// 			equal := actual.PublishedAt.Equal(tc.expected.PublishedAt.Time)
// 			if equal != tc.equal {
// 				t.Errorf("%s: actual=%#v, expected=%#v, equal=%v, expected=%v", tc.desc, actual, tc.expected, equal, tc.equal)
// 				continue
// 			}
// 		}
// 	}
// }
