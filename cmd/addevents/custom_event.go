package main

type CustomEvent struct {
	session Session
	name    string
	keys    []string
	values  []string
}

// Row encode session as a flat []any slice ready to be sent to ClickHouse.
func (ce *CustomEvent) Row() (record []any) {
	record = append(record,
		ce.session.exitTimestamp,
		ce.session.domain,
		ce.session.exitPath,
		ce.session.visitorId,
		ce.session.sessionUuid,
		ce.name,
		ce.keys,
		ce.values,
	)

	return
}
