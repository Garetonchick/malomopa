package common

type Order struct {
	OrderID    string
	ExecutorID string
	Cost       float32
	Payload    OrderPayload
}

type OrderPayload []byte
type OrderInfo map[string]any

type GeneralOrderInfo struct {
	ID             string  `json:"id"`
	UserID         string  `json:"user_id"`
	ZoneID         string  `json:"zone_id"`
	BaseCoinAmount float32 `json:"base_coin_amount"`
}

type ZoneInfo struct {
	ID          string  `json:"id"`
	CoinCoeff   float32 `json:"coin_coeff"`
	DisplayName string  `json:"display_name"`
}

type ExecutorProfile struct {
	ID     string   `json:"id"`
	Tags   []string `json:"tags"`
	Rating float32  `json:"rating"`
}

type TollRoadsInfo struct {
	BonusAmount float32 `json:"bonus_amount"`
}
