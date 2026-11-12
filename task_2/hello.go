package main 
import (
	"fmt"
	"log"
)

func main() {
	ba := BankAccount{
		id : 1,
		name : "jhon",
		balance : 12000.00,
	}
	// ba.GetInfo()
	// err, val := ba.Withdraw(-1)
	// if err != nil {
	// 	log.Println(err)
	// }else{
	// 	log.Println(val)
	// }

	ba.GetInfo()
	depVal, depErr := ba.Deposit(1000)
	if depErr != nil {
		log.Fatal(depErr)
	}else{
		log.Println(depVal)
	}
	ba.GetInfo()
}

type BankAccount struct {
	id int 
	name string
	balance float64
}

func (b *BankAccount) Balance() float64{
	return b.balance
}

func (b *BankAccount) GetInfo()  {
	fmt.Printf("Account Name : %v\nBalance : %v\n", b.name, b.balance)
}

func (b *BankAccount) Withdraw(money float64) (error, string) {
	// edge case
	if money < 0 || money == 0.0{
		return fmt.Errorf("Invalid ammount"), "Invalid ammount"
	}
	if money > b.balance {
		return fmt.Errorf("Insufficient Balance"),"Insufficient Balance"
	}else{
		b.balance -= money
		return nil, "Money withdrawed successfully"
	}
}

func (b *BankAccount) Deposit(money float64) (string, error){
	// edge case
	if money <= 0 {
		return  ("Cannot deposit 0  money"),fmt.Errorf("Money cannot be 0.0")
	}
	b.balance += money 
	return fmt.Sprintf("Successfully deposited money. Current Balance : %v", b.Balance()),  nil
}



