package shared

type Transaction struct {
	MTI string `json:"mti"`
	F2  string `json:"f2"`  // card number
	F3  string `json:"f3"`  // card expiry
	F4  string `json:"f4"`  // amount
	F12 string `json:"f12"` // local transaction time
	F13 string `json:"f13"` // local transaction date
	F38 string `json:"f38"` // authorization code response
	F39 string `json:"f39"` // response code
}
