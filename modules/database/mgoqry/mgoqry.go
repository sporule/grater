package mgoqry

import "go.mongodb.org/mongo-driver/bson"

//Bson returns a bson.M key value pair
func Bson(key string, value interface{}) bson.M {
	return bson.M{key: value}
}

//Bsons returns multiple bson.M key value pairs
func Bsons(keyValuePairs map[string]interface{}) bson.M {
	qry := bson.M{}
	if keyValuePairs != nil {
		for key, value := range keyValuePairs {
			qry[key] = value
		}
	}
	return qry
}

//All will match all queies in arrary
func All(values ...interface{}) bson.M {
	return Bson("$all", values)
}

//In will match any queries in arrary
func In(values ...interface{}) bson.M {
	return Bson("$in", values)
}

//Nin will match anything other than the queies in arrary
func Nin(values ...interface{}) bson.M {
	return Bson("$nin", values)
}

//Eq matches equale comparison
func Eq(value interface{}) bson.M {
	return Bson("$eq", value)
}

//Gt matches greater comparison
func Gt(value interface{}) bson.M {
	return Bson("$gt", value)
}

//Gte matches greater or equal comparison
func Gte(value interface{}) bson.M {
	return Bson("$gte", value)
}

//Lt matches less comparison
func Lt(value interface{}) bson.M {
	return Bson("$lt", value)
}

//Lte matches less or equal comparison
func Lte(value interface{}) bson.M {
	return Bson("$lte", value)
}

//And provides and relationship
func And(queries ...interface{}) bson.M {
	return Bson("$and", queries)
}

//Or provides and relationship
func Or(values ...interface{}) bson.M {
	return Bson("$Or", values)
}

//Not provides NOT relationship
func Not(value interface{}) bson.M {
	return Bson("$not", value)
}

//Nor provides NOR relationship
func Nor(values ...interface{}) bson.M {
	return Bson("$nor", values)
}

//Select takes fields name and returns the "filenames":"1" to select the input fields
func Select(isSelect bool, values ...string) bson.M {
	selector := bson.M{}
	for _, value := range values {
		selector[value] = isSelect
	}
	return selector
}
