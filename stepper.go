package main

type Stepper interface {
	Next() Item
	Peek() Item
	ConsumeUntil(itemType ItemType)
	Errorf()
}
