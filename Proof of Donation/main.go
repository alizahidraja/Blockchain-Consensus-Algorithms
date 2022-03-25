	package main

	import (
		"crypto/sha256"
		"encoding/gob"
		"fmt"
		"log"
		"net"
		"os"
		"strconv"
	)


type Block struct {
	//Hash
	//Data

	Transaction string
	Reward string
	PrevPointer *Block
	Hash [32]byte
	PrevHash [32]byte
}

type Person struct{

	Name string
	Wallet float64
	Port string
	MinePort string
	DonationPort string
	SenderPort string
	Donation float64

}

type Transaction struct{

	Sender int
	Amount float64
	Receiver int
	Miner int
	Transfee float64
	Donation float64

}


func  DeriveHash(Transaction string)[32]byte {
	return sha256.Sum256([]byte(Transaction))
}


func InsertBlock(Transaction,Reward string, chainHead *Block) *Block {
	if chainHead == nil {
		return &Block{Transaction,Reward, nil,DeriveHash(Transaction+"+"+Reward),[32]byte{}}
	}
	return &Block{Transaction,Reward, chainHead,DeriveHash(Transaction+"+"+Reward),DeriveHash(chainHead.Transaction+"+"+chainHead.Reward)}

}

func ListBlocks(chainHead *Block) {
	for p := chainHead; p != nil; p = p.PrevPointer {
		fmt.Printf("Transaction: %s, Hash:%x, PrevHash:%x\n",p.Transaction+" & "+p.Reward,p.Hash,p.PrevHash)

	}
}




func VerifyChain(chainHead *Block) {
	for p := chainHead; p != nil; p = p.PrevPointer {
		if p.PrevPointer !=nil{
			if p.PrevHash != p.PrevPointer.Hash{
				fmt.Println("Chain is InvMDd!")
				return
			}
		}
	}
	fmt.Println("Chain is VMDd!")
}



func ExecuteTransactionToAddPerson(Sender *Person,SendingAmount float64,Receiver,Miner *Person,TransationFee,MinerReward float64,chainHead *Block)*Block{
	if Sender.Wallet>=SendingAmount{
		if TransationFee<=SendingAmount {
			Sender.Wallet -= SendingAmount
			Receiver.Wallet += SendingAmount - TransationFee
			Miner.Wallet += MinerReward+TransationFee


			return InsertBlock(Sender.Name+"-"+fmt.Sprintf("%f", SendingAmount)+" ==> "+fmt.Sprintf("%f", SendingAmount- TransationFee)+"->"+Receiver.Name, fmt.Sprintf("%f", MinerReward+TransationFee)+"->"+Miner.Name, chainHead)

		}else{
			fmt.Println("InvMDd Reward Amount!")
		}
	}else{
		fmt.Println(Sender.Name,"does not have enough BC!")
	}
	return chainHead
}

func ExecuteTransaction(Sender *Person,SendingAmount float64,Receiver,Miner *Person,TransationFee,MinerReward float64,chainHead *Block,rewards string)*Block{
		if Sender.Wallet>=SendingAmount{
			if TransationFee<=SendingAmount {
				Sender.Wallet -= SendingAmount
				Receiver.Wallet += SendingAmount - TransationFee
				Miner.Wallet += MinerReward+TransationFee
				Donation:=Miner.Donation
				Miner.Wallet-=Donation
				Miner.Donation=0
				return InsertBlock(Sender.Name+"-"+fmt.Sprintf("%f", SendingAmount)+" "+Miner.Name+"-"+fmt.Sprintf("%f", Donation)+" ==> "+fmt.Sprintf("%f", SendingAmount- TransationFee)+"->"+Receiver.Name, fmt.Sprintf("%f", MinerReward+TransationFee)+"->"+Miner.Name+" "+rewards, chainHead)

			}else{
				fmt.Println("Invalid Reward Amount!")
			}
		}else{
			fmt.Println(Sender.Name,"does not have enough BC!")
		}
		return chainHead
}

func ReceiveUpdatedBlockChain(ln net.Listener,chainHead *Block,people []Person) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		dec := gob.NewDecoder(conn)
		err = dec.Decode(&chainHead)
		if err != nil {
			log.Println(err)
		}
		err = dec.Decode(&people)
		if err != nil {
			log.Println(err)
		}

		fmt.Println("Latest Blockchain and Ledger!")
		ListBlocks(chainHead)
		fmt.Println(people)
		fmt.Println("1. Do you want to make a transaction?\n2. Do you want to exit?")

	}
}

func BroadCastBlockChainAndLedger(chainHead *Block,people []Person,ind int){
	for i:=0;i<len(people);i++ {
		if i != ind {
			fmt.Println("Sending BlockChain & Ledger to " + people[i].Name)
			conn, err := net.Dial("tcp", "localhost:"+people[i].Port)
			if err != nil {
				//handle error
				fmt.Println(err)
			}
			gobEncoder := gob.NewEncoder(conn)
			err = gobEncoder.Encode(chainHead)
			//gob.RegisterName()
			if err != nil {
				//handle error
				fmt.Println(err)
			}
			err = gobEncoder.Encode(people)
			if err != nil {
				//handle error
				fmt.Println(err)
			}
		}
	}
}

func MinerHandler(ln net.Listener,chainHead *Block,people []Person,minerPort string, ind int) {
	ln, err := net.Listen("tcp", ":"+minerPort)
	if err != nil {
		log.Fatal(err)
	}
	MinerReward:=100.0
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}
		var transaction Transaction
		gobEncoder := gob.NewDecoder(conn)
		err = gobEncoder.Decode(&transaction)
		fmt.Println(transaction)
		rewards:=""
		for i:=0;i<len(people);i++{
			if i!=transaction.Miner{
				people[i].Donation=0
				number:=float64(len(people)-1)
				people[i].Wallet+=people[transaction.Miner].Donation*(1/number)
				rewards+=fmt.Sprintf("%f", people[transaction.Miner].Donation*(1/number))+"->"+people[i].Name+" "
			}
		}
		chainHead = ExecuteTransaction(&people[transaction.Sender],transaction.Amount,&people[transaction.Receiver],&people[transaction.Miner],transaction.Transfee,MinerReward,chainHead,rewards)

		fmt.Println("Latest Blockchain and Ledger!")
		ListBlocks(chainHead)
		fmt.Println(people)
		fmt.Println("BroadCasting!")
		BroadCastBlockChainAndLedger(chainHead,people,ind)
		fmt.Println("1. Do you want to make a transaction?\n2. Do you want to exit?")
	}
}

func DonationHandler(ln net.Listener,people []Person,donationPort string, ind int){
	ln, err := net.Listen("tcp", ":"+donationPort)
	if err != nil {
		log.Fatal(err)
	}
	var sender string
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}
		gobDecoder := gob.NewDecoder(conn)
		err = gobDecoder.Decode(&sender)
		fmt.Println("A block is up for mining, how much do you want to donate(The person with the most donation gets to mine)")
		var donationAmount float64
		_, err = fmt.Scanln( &donationAmount)
		if err!=nil{
			fmt.Println(err)
		}

		people[ind].Donation=donationAmount

		conn, err = net.Dial("tcp", "localhost:"+sender)
		if err != nil {
			//handle error
			fmt.Println(err)
		}
		gobEncoder := gob.NewEncoder(conn)
		err = gobEncoder.Encode(people)
		fmt.Println("Sent Updated Donation Amount")
	}


}
func BroadCastDonation(people []Person,ind int){
	for i:=0;i<len(people);i++ {
		if i != ind {
			conn, err := net.Dial("tcp", "localhost:"+people[i].DonationPort)
			if err != nil {
				//handle error
				fmt.Println(err)
			}
			gobEncoder := gob.NewEncoder(conn)
			err = gobEncoder.Encode(people[ind].SenderPort)
		}
	}
}

func main() {


	fmt.Println("Welcone to MDCoin (MDC), The developer of this Currency is MD")
	MinerReward:=100.0
	if len(os.Args) == 2{//Satoshi
		//max users=os.args[1]
		Users,_:=strconv.Atoi(os.Args[1])
		var chainHead *Block

		people:=make([]Person,0)
		people = append(people,Person{"MD",100,"6000","5000","4000","3000",0} )


		fmt.Println(people)

		fmt.Println("This is MD, making the genesis Block")

		chainHead = InsertBlock("GenesisBlock", "100->MD",nil)


		fmt.Println(people)
		ListBlocks(chainHead)
		//ChangeBlock("MDceToBob", "MDceToTrudy", chainHead)
		VerifyChain(chainHead)
		//		fmt.Println(len(os.Args), os.Args)



		ln, err := net.Listen("tcp", ":6000")
		if err != nil {
			log.Fatal(err)
		}
		for {//Waiting for All Users
			fmt.Println("Waiting for Users")
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			var person Person
			gobDecoder:=gob.NewDecoder(conn)
			err=gobDecoder.Decode(&person)
			fmt.Println("New Person "+person.Name+" added!")
			people=append(people,person)
			chainHead = ExecuteTransactionToAddPerson(&people[len(people)-1],0,&people[len(people)-1],&people[0],0,MinerReward,chainHead)
			if err != nil {
				log.Println(err)
			}
			Users=Users-1
			if Users==0{
				fmt.Println("Max users have entered the network, starting propagation!")
				break
			}
		}
		fmt.Println(people)

		for i:=1;i<len(people);i++ {

			fmt.Println("Sending BlockChain & Ledger to " + people[i].Name)
			conn, err := net.Dial("tcp", "localhost:"+people[i].Port)
			if err != nil {
				//handle error
				fmt.Println(err)
			}
			gobEncoder := gob.NewEncoder(conn)
			err=gobEncoder.Encode(i)
			err = gobEncoder.Encode(chainHead)
			//gob.RegisterName()
			if err != nil {
				//handle error
				fmt.Println(err)
			}
			err = gobEncoder.Encode(people)
			if err != nil {
				//handle error
				fmt.Println(err)
			}

		}
		fmt.Println("Latest Blockchain!")
		ListBlocks(chainHead)

		go ReceiveUpdatedBlockChain(ln,chainHead,people)
		go MinerHandler(ln,chainHead,people,"5000",0)
		go DonationHandler(ln,people,"4000",0)
		Senderln, err := net.Listen("tcp", ":3000")
		if err != nil {
			log.Fatal(err)
		}
		for{
			fmt.Println("1. Do you want to make a Transaction?\n2. Do you want to exit?")
			var i,j int
			var amount,tranfee float64
			_, err := fmt.Scanln(&i)
			if err!=nil{
				fmt.Println(err)
			}
			if i == 1{//Transaction
				fmt.Println("Who do you want to send BC to?")
				for i=0;i<len(people);i++{
					if i!=0 {
						fmt.Println(i,people[i].Name)
					}
				}
				_, err := fmt.Scanln( &j)
				if err!=nil{
					fmt.Println(err)
				}

				fmt.Println("How much do you want to send to",people[j].Name,"?")
				_, err = fmt.Scanln( &amount)
				if err!=nil{
					fmt.Println(err)
				}

				fmt.Println("How much do you wanna give as Transaction fees ?")
				_, err = fmt.Scanln( &tranfee)
				if err!=nil{
					fmt.Println(err)
				}


				//BroadCast
				BroadCastDonation(people,0)
				fmt.Println("A block is up for mining, how much do you want to donate(The person with the most donation gets to mine)")
				var donationAmount float64
				_, err = fmt.Scanln( &donationAmount)
				if err!=nil{
					fmt.Println(err)
				}
				people[0].Donation=donationAmount



				for i=0;i<len(people);i++{
					if i!=0{
						fmt.Println("Waiting for responses at")
						conn, err := Senderln.Accept()
						if err != nil {
							log.Println(err)
						}
						gobEncoder := gob.NewDecoder(conn)
						err = gobEncoder.Decode(&people)
						if err != nil {
							log.Println(err)
						}
					}

				}
//				time.Sleep(20*time.Second)
				d:=0.0
				miner:=0
				for i=0;i<len(people);i++{
					if d<people[i].Donation{
						d=people[i].Donation
						miner=i
					}
				}


				fmt.Println(miner,people[miner].Name,"Has given the most donation and is selected as miner")

				fmt.Println("Sending Transaction to Miner, i.e.",people[miner].Name)
				conn, err := net.Dial("tcp", "localhost:"+people[miner].MinePort)
				if err != nil {
					//handle error
					fmt.Println(err)
				}
				gobEncoder := gob.NewEncoder(conn)
				err = gobEncoder.Encode(Transaction{0,amount,j,miner,tranfee,people[miner].Donation})


			}else if i ==2{
				break
			}
		}

	}else {//Peers
		//arg[1]=Name, arg[2]=Port, arg[3]=MinePort, arg[4]=DonationPort, arg[5]=SenderPort
		fmt.Println("This is a new member of the block, named ",os.Args[1])
		conn, err := net.Dial("tcp", "localhost:6000")
		if err != nil {
			//handle error
		}
		person:=&Person{os.Args[1],100,os.Args[2],os.Args[3],os.Args[4],os.Args[5],0}
		enc :=gob.NewEncoder(conn)
		err=enc.Encode(person)




		ln, err := net.Listen("tcp", ":"+os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		conn, err = ln.Accept()
		if err != nil {
			log.Println(err)
		}
		var ind int
		var chainHead *Block
		dec := gob.NewDecoder(conn)
		err=dec.Decode(&ind)

		err = dec.Decode(&chainHead)
		if err != nil {
			log.Println(err)
		}
		var people []Person
		err = dec.Decode(&people)
		if err != nil {
			log.Println(err)
		}

		ListBlocks(chainHead)
		fmt.Println(people)

		go ReceiveUpdatedBlockChain(ln,chainHead,people)
		go MinerHandler(ln,chainHead,people,os.Args[3],ind)
		go DonationHandler(ln,people,os.Args[4],ind)
		Senderln, err := net.Listen("tcp",":"+ os.Args[5])
		if err != nil {
			log.Fatal(err)
		}
		for{
			fmt.Println("1. Do you want to make a Transaction?\n2. Do you want to exit?")
			var i,j int
			var amount,tranfee float64
			_, err := fmt.Scanln(&i)
			if err!=nil{
				fmt.Println(err)
			}
			if i == 1{//Transaction
				fmt.Println("Who do you want to send BC to?")
				for i=0;i<len(people);i++{
					if i!=ind {
						fmt.Println(i,people[i].Name)
					}
				}
				_, err := fmt.Scanln( &j)
				if err!=nil{
					fmt.Println(err)
				}

				fmt.Println("How much do you want to send to",people[j].Name,"?")
				_, err = fmt.Scanln( &amount)
				if err!=nil{
					fmt.Println(err)
				}

				fmt.Println("How much do you wanna give as Transaction fees ?")
				_, err = fmt.Scanln( &tranfee)
				if err!=nil{
					fmt.Println(err)
				}


				//Broadcast Mining alert
				BroadCastDonation(people,ind)
				fmt.Println("A block is up for mining, how much do you want to donate(The person with the most donation gets to mine)")
				var donationAmount float64
				_, err = fmt.Scanln( &donationAmount)
				if err!=nil{
					fmt.Println(err)
				}
				people[ind].Donation=donationAmount

				for i=0;i<len(people);i++{
					if i!=ind{
						fmt.Println("Waiting for responses at")
						conn, err := Senderln.Accept()
						if err != nil {
							log.Println(err)
						}
						gobEncoder := gob.NewDecoder(conn)
						err = gobEncoder.Decode(&people)
						if err != nil {
							log.Println(err)
						}
					}

				}
				//				time.Sleep(20*time.Second)
				d:=0.0
				miner:=0
				for i=0;i<len(people);i++{
					if d<people[i].Donation{
						d=people[i].Donation
						miner=i
					}
				}



				fmt.Println(miner,people[miner],"is selected as miner")
				fmt.Println("Sending Transaction to Miner, i.e.",people[miner])
				conn, err := net.Dial("tcp", "localhost:"+people[miner].MinePort)
				if err != nil {
					//handle error
					fmt.Println(err)
				}
				gobEncoder := gob.NewEncoder(conn)
				err = gobEncoder.Encode(Transaction{ind,amount,j,miner,tranfee,people[miner].Donation})


			}else if i ==2{
				break
			}
		}

	}

}

