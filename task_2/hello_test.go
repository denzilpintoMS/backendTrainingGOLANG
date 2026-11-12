package main

import ("testing")

// func TestWithDraw(t *testing.T){
// 	dummyAcc := BankAccount{
// 		id : 2,
// 		name : "doe",
// 		balance : 0.0,
// 	}
// }

func TestDeposit(t *testing.T){
tests := []struct {
    name          string
    balance       float64
    deposit       float64
    wantedBalance float64
    wantErr       bool
}{
    {"Valid deposit", 0, 1000, 1000, false},
    {"Deposit into existing balance", 5000, 1000, 6000, false},
    {"Negative deposit", 5000, -1000, 5000, true},
    {"Zero deposit", 5000, 0, 5000, true},
    {"Small positive deposit", 5000, 0.50, 5000.50, false},
}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			acc := BankAccount{
				balance: float64(tt.balance),
			}

			_, err := acc.Deposit(float64(tt.deposit))
			if (err != nil) != tt.wantErr{
				t.Errorf("got error %v want err %v", err, tt.wantErr)
			}

			gotBaln := acc.Balance()
			if gotBaln != tt.wantedBalance {
				t.Errorf("got balance %v want error %v",gotBaln, tt.wantedBalance)
		}
		})
	}
}
func TestWithdraw(t *testing.T){
    tests := []struct {
        name        string
        balance     float64
        withdraw    float64
        wantBalance float64
        wantErr     bool
    }{
        {"Valid withdrawal", 5000, 1000, 4000, false},
        {"Insufficient balance", 500, 1000, 500, true},
        {"Negative withdrawal", 5000, -100, 5000, true},
        {"Zero withdrawal", 5000, 0, 5000, true},
        {"Withdraw full balance", 5000, 5000, 0, false},
    }
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			acc := BankAccount{
				balance: float64(tt.balance),
			}

			err, _ := acc.Withdraw(float64(tt.withdraw))
			if (err != nil) != tt.wantErr{
				t.Errorf("got error %v want err %v", err, tt.wantErr)
			}

			gotBaln := acc.Balance()
			if gotBaln != tt.wantBalance {
				t.Errorf("got balance %v want error %v",gotBaln, tt.wantBalance)
		}
		})
	}
}

// func TestBalance(t *testing.T){
// 	tests := []struct {
//
// 	}
// }
