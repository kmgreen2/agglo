package search

type ElasticIndexValue struct {
	values map[string]*IndexItem
	id string
}

func (indexValue *ElasticIndexValue) Values() map[string]*IndexItem {
	return indexValue.values
}

func (indexValue *ElasticIndexValue) Id() string {
	return indexValue.id
}

type ElasticSearchResults struct {
	values []*ElasticIndexValue
}

type ElasticIndexValueBuilder struct {
	value *ElasticIndexValue
}

func NewElasticIndexValueBuilder(id string) *ElasticIndexValueBuilder {
	return &ElasticIndexValueBuilder{value: &ElasticIndexValue{
		values: make(map[string]*IndexItem),
		id: id,
	}};
}

func (builder *ElasticIndexValueBuilder) AddNumeric(key string, value float64) *ElasticIndexValueBuilder {
	builder.value.values[key] = &IndexItem{
		itemType: IndexItemNumeric,
		item: value,
	}
	return builder
}

func (builder *ElasticIndexValueBuilder) AddKeyword(key string, value string) *ElasticIndexValueBuilder {
	builder.value.values[key] = &IndexItem{
		itemType: IndexItemKeyword,
		item: value,
	}
	return builder
}

func (builder *ElasticIndexValueBuilder) AddFreeText(key string, value string) *ElasticIndexValueBuilder {
	builder.value.values[key] = &IndexItem{
		itemType: IndexItemFreeText,
		item: value,
	}
	return builder
}

func (builder *ElasticIndexValueBuilder) AddDate(key string, value int64) *ElasticIndexValueBuilder {
	builder.value.values[key] = &IndexItem{
		itemType: IndexItemDate,
		item: value,
	}
	return builder
}

func (builder *ElasticIndexValueBuilder) SetBlob(value []byte) *ElasticIndexValueBuilder {
	builder.value.values["_blob"] = &IndexItem{
		itemType: IndexItemBlob,
		item: value,
	}
	return builder
}

func (builder *ElasticIndexValueBuilder) Get() *ElasticIndexValue {
	return builder.value
}

func ResolveElasticResults(rawResults map[string]interface{}) (ElasticSearchResults, error) {
	var results ElasticSearchResults

	if hitsResults, ok := rawResults["hits"].(map[string]interface{}); ok {
		if hitsList, okList := hitsResults["hits"].([]interface{}); okList {
			for _, hit := range hitsList {
				if hitMap, okHit := hit.(map[string]interface{}); okHit {
					if hitEntry, okHitEntry := hitMap["_source"].(map[string]interface{}); okHitEntry {
						builder := NewElasticIndexValueBuilder("")
						for entryType, entry := range hitEntry {
							if values, valuesOK := entry.([]interface{}); valuesOK {
								for _, v := range values {
									switch entryType {
									case ElasticNumeric:
										if entryValue, entryValueOk := v.(map[string]interface{}); entryValueOk {
											builder.AddNumeric(entryValue["id"].(string), entryValue["value"].(float64))
										}
										break
									case ElasticKeyword:
										if entryValue, entryValueOk := v.(map[string]interface{}); entryValueOk {
											builder.AddKeyword(entryValue["id"].(string), entryValue["value"].(string))
										}
										break
									case ElasticFreeText:
										if entryValue, entryValueOk := v.(map[string]interface{}); entryValueOk {
											builder.AddFreeText(entryValue["id"].(string), entryValue["value"].(string))
										}
										break
									case ElasticDate:
										if entryValue, entryValueOk := v.(map[string]interface{}); entryValueOk {
											builder.AddDate(entryValue["id"].(string), entryValue["value"].(int64))
										}
										break
									case ElasticBlob:
										// ToDo(KMG): Need to deser the base64 results and store as []byte
										builder.SetBlob([]byte(v.(string)))
										break
									}
								}
							} else {
							}
						}
						results.values = append(results.values, builder.Get())
					} else {
					}
				} else {
				}
			}
		} else {
		}
	} else {
	}
	return results, nil
}
