package database

import "reflect"

func QueryDB(interchangeableStruct reflect.Type, query string, args ...any) ([]any, error) {
	var structArray []any
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		newStructPointer := reflect.New(interchangeableStruct)
		structElement := newStructPointer.Elem()

		numCols := structElement.NumField()
		columns := make([]any, numCols)

		for i := range numCols {
			field := structElement.Field(i)
			columns[i] = field.Addr().Interface()
		}

		if err := rows.Scan(columns...); err != nil {
			return nil, err
		}

		structArray = append(structArray, structElement.Interface())
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return structArray, nil
}
