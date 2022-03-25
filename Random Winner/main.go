package main

import (
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
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
	PoolPort string
	SenderPort string
}

type Transaction struct{

	Sender int
	Amount float64
	Receiver int
	Miner int
	Transfee float64
	Pool PrizePool

}

type PrizePool struct{
	Prize float64
	MinEntry float64
	MaxEntry float64
	Miners []int
	Entries []float64
	Winner int
	SenderPort string
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




/*
func ChangeBlock(oldTrans string, newTrans string, chainHead *Block) {
	for p := chainHead; p != nil; p = p.next {
		if p.Transaction == oldTrans{
			p.Transaction=newTrans
			p.Hash=DeriveHash(newTrans)

		}
	}
}
*/
func VerifyChain(chainHead *Block) {
	for p := chainHead; p != nil; p = p.PrevPointer {
		if p.PrevPointer !=nil{
			if p.PrevHash != p.PrevPointer.Hash{
				fmt.Println("Chain is Invalid!")
				return
			}
		}
	}
	fmt.Println("Chain is Valid!")
}



func handleConnection(c net.Conn) {
	log.Println("A client has connected", c.RemoteAddr())
	c.Write([]byte("Hello world"))
}

func ExecuteTransactionforAddingMembers(Sender *Person,SendingAmount float64,Receiver,Miner *Person,TransationFee,MinerReward float64,chainHead *Block)*Block{
	if Sender.Wallet>=SendingAmount{
		if TransationFee<=SendingAmount {
			Sender.Wallet -= SendingAmount
			Receiver.Wallet += SendingAmount - TransationFee
			Miner.Wallet += MinerReward+TransationFee
			return InsertBlock(Sender.Name+"-"+fmt.Sprintf("%f", SendingAmount)+" ==> "+fmt.Sprintf("%f", SendingAmount- TransationFee)+"->"+Receiver.Name, fmt.Sprintf("%f", MinerReward+TransationFee)+"->"+Miner.Name, chainHead)

		}else{
			fmt.Println("Invalid Reward Amount!")
		}
	}else{
		fmt.Println(Sender.Name,"does not have enough BC!")
	}
	return chainHead
}

func ExecuteTransaction(Sender *Person,SendingAmount float64,Receiver,Miner *Person,TransationFee float64,Pool PrizePool,chainHead *Block,people []Person)*Block{
	if Sender.Wallet>=SendingAmount{
		if TransationFee<=SendingAmount {
			Sender.Wallet -= SendingAmount
			Receiver.Wallet += SendingAmount - TransationFee
			Reward:=Pool.Prize/2+TransationFee
			Miner.Wallet += Reward

			entries:=" "
			Entry:=0.0
			for i:=0;i<len(Pool.Miners);i++ {
				if Pool.Miners[i]!=Pool.Winner{
					Entry+=Pool.Entries[i]
				}
				people[Pool.Miners[i]].Wallet-=Pool.Entries[i]
				entries+=people[Pool.Miners[i]].Name+"-"+fmt.Sprintf("%f", Pool.Entries[i])+" "
			}
            rewards:=" "
			for i:=0;i<len(Pool.Miners);i++ {
				if Pool.Miners[i]!=Pool.Winner{
					people[Pool.Miners[i]].Wallet+=(Pool.Entries[i]/Entry)*(Pool.Prize/2)
					rewards+=fmt.Sprintf("%f",(Pool.Entries[i]/Entry)*(Pool.Prize/2))+"->"+people[Pool.Miners[i]].Name+" "
				}
			}
			return InsertBlock(Sender.Name+"-"+fmt.Sprintf("%f", SendingAmount)+entries+" ==> "+fmt.Sprintf("%f", SendingAmount- TransationFee)+"->"+Receiver.Name, fmt.Sprintf("%f", Reward)+"->"+Miner.Name+rewards, chainHead)

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


func ReceivePrizePool(ln net.Listener,PoolPort string,ind int,people []Person) {
	ln, err := net.Listen("tcp", ":"+PoolPort)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		var pool PrizePool
		dec := gob.NewDecoder(conn)
		err = dec.Decode(&pool)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("PrizePool Information!")
		fmt.Println(pool)

		//fmt.Println("1. Do you want to make a transaction?\n2. Do you want to exit?")
		fmt.Println("Minimum Entry for the pool is",pool.MinEntry,"And the Maximum Entry is",pool.MaxEntry)
		var e int
		var entry float64
		_, err = fmt.Scanln(&e)
		if err!=nil{
			fmt.Println(err)
		}
		fmt.Println("How much do you want to give in the pool?")
		_, err = fmt.Scanln( &entry)
		if err!=nil{
			fmt.Println(err)
		}
		if entry>=pool.MinEntry{
			if entry<=pool.MaxEntry{
				if entry<=people[ind].Wallet{
					pool.Prize+=entry
					pool.Entries = append(pool.Entries, entry)
					pool.Miners = append(pool.Miners, ind)
					//Broadcast Pool
				}else{
					fmt.Println("You donot have",entry,"in your wallet")
				}
			}else{
				fmt.Println("Invalid!")
			}
		}else{
			fmt.Println("Invalid!")
		}
		//BroadCastPrizePool(pool,people,ind)
		SendPrizePoolToSender(pool)


	}
}

func SendPrizePoolToSender(pool PrizePool){
	conn, err := net.Dial("tcp", "localhost:"+pool.SenderPort)
	if err != nil {
		//handle error
		fmt.Println(err)
	}
	gobEncoder := gob.NewEncoder(conn)
	err = gobEncoder.Encode(pool)
	fmt.Println("Sent Updated Pool")

}

func BroadCastPrizePool(pool PrizePool,people []Person,ind,i int) {

	fmt.Println("Sending PrizePool Information " + people[i].Name)
	conn, err := net.Dial("tcp", "localhost:"+people[i].PoolPort)
	if err != nil {
		//handle error
		fmt.Println(err)
	}
	gobEncoder := gob.NewEncoder(conn)
	err = gobEncoder.Encode(&pool)
	//gob.RegisterName()
	if err != nil {
		//handle error
		fmt.Println(err)
	}
}

func MinerHandler(ln net.Listener,chainHead *Block,people []Person,minerPort string, ind int) {
	ln, err := net.Listen("tcp", ":"+minerPort)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}
		var transaction Transaction
		gobEncoder := gob.NewDecoder(conn)
		err = gobEncoder.Decode(&transaction)
		fmt.Println(transaction)
		chainHead = ExecuteTransaction(&people[transaction.Sender],transaction.Amount,&people[transaction.Receiver],&people[transaction.Miner],transaction.Transfee,transaction.Pool,chainHead,people)
		fmt.Println("Latest Blockchain and Ledger!")
		ListBlocks(chainHead)
		fmt.Println(people)
		fmt.Println("BroadCasting!")
		BroadCastBlockChainAndLedger(chainHead,people,ind)
		fmt.Println("1. Do you want to make a transaction?\n2. Do you want to exit?")
	}
}


func main() {
	/*
		fmt.Println(reflect.TypeOf(tst))
		tst1 := &Person{"",0.0,""}

		fmt.Println(reflect.TypeOf(tst1))
		fmt.Println(reflect.TypeOf(tst1)==reflect.TypeOf(&Person{"",0,""}))
	*/

	fmt.Println("Welcone to BlockCoin (BC), The developer of this Currency is Ali")

	if len(os.Args) == 2{//Satoshi
		//max users=os.args[1]
		Users,_:=strconv.Atoi(os.Args[1])
		var chainHead *Block


		//var max_nodes int
		//_, _ = fmt.Scan(&age)

		//reader := bufio.newReader(os.Stdin)
		//var name string
		//fmt.Println("What is your name?")
		//name, _ := reader.readString("\n")

		//max_nodes = 6
		//_, _ = fmt.Scan(&max_nodes)
		//var nodes int = 0

		people:=make([]Person,0)
		people = append(people,Person{"Ali",100,"6000","5000","4000","3000"} )
		//people = append(people,Person{"Dani",0,6001} )
		//people = append(people,Person{"Saad",0,6001} )
		//people = append(people,Person{"Shah",0,6001} )
		//people = append(people,Person{"Joan",0,6001} )
		//people = append(people,Person{"Rajja",0,6001} )

		fmt.Println(people)

		fmt.Println("This is Ali, making the genesis Block")

		chainHead = InsertBlock("GenesisBlock", "100->Ali",nil)

		//chainHead = ExecuteTransaction(&people[0],2,&people[1],&people[2],0.2,chainHead)
		//chainHead = ExecuteTransaction(&people[1],2,&people[3],&people[4],0,chainHead)
		//chainHead = ExecuteTransaction(&people[0],2,&people[1],&people[2],0.2,chainHead)
		//chainHead = ExecuteTransaction(&people[1],2,&people[3],&people[4],0,chainHead)
		//chainHead = ExecuteTransaction(&people[0],2,&people[4],&people[5],0.2,chainHead)
		//chainHead = ExecuteTransaction(&people[0],2,&people[5],&people[4],0.2,chainHead)

		fmt.Println(people)
		ListBlocks(chainHead)
		//ChangeBlock("AliceToBob", "AliceToTrudy", chainHead)
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
			//gobEncoder := gob.NewEncoder(conn)
			//err = gobEncoder.Encode(chainHead)
			var person Person
			gobDecoder:=gob.NewDecoder(conn)
			err=gobDecoder.Decode(&person)
			fmt.Println("New Person "+person.Name+" added!")
			people=append(people,person)
			chainHead = ExecuteTransactionforAddingMembers(&people[len(people)-1],0,&people[len(people)-1],&people[0],0,100.0,chainHead)
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
		ind:=0
		go ReceiveUpdatedBlockChain(ln,chainHead,people)
		go ReceivePrizePool(ln,"4000",0,people)
		go MinerHandler(ln,chainHead,people,"5000",0)
		poolln, err := net.Listen("tcp",":3000")
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

				//New Algorithm
				var pool PrizePool
				pool.Prize=0.0
				fmt.Println("Your transaction is pending to be made part of the blockchain! A prize pool with Minimum and Maximum values is going to be set!")
				pool.MinEntry= float64(rand.Intn(51))
				pool.MaxEntry=pool.MinEntry*2
				pool.SenderPort=people[0].SenderPort
				//Broadcast Pool
				//BroadCastPrizePool(pool,people,ind)
				//
				fmt.Println("Minimum Entry for the pool is",pool.MinEntry,"And the Maximum Entry is",pool.MaxEntry)

				var entry float64
					fmt.Println("How much do you want to give in the pool?")
					_, err = fmt.Scanln( &entry)
					if err!=nil{
						fmt.Println(err)
					}
				if entry>=pool.MinEntry{
					if entry<=pool.MaxEntry{
						if entry<=people[ind].Wallet{
							people[ind].Wallet-=entry
							pool.Prize+=entry
							pool.Entries = append(pool.Entries, entry)
							pool.Miners = append(pool.Miners, ind)
							//Broadcast Pool

						}else{
							fmt.Println("You donot have",entry,"in your wallet")
						}
					}else{
						fmt.Println("Invalid!")
					}
				}else{
					fmt.Println("Invalid!")
				}
				//pool.Miners
				//Recieve the final prize pool
				for i=0;i<len(people);i++{
					if i!=ind{
						BroadCastPrizePool(pool,people,ind,i)
						fmt.Println("Waiting for the updated pool")
						conn, err := poolln.Accept()
						if err != nil {
							log.Println(err)
						}
						gobEncoder := gob.NewDecoder(conn)
						err = gobEncoder.Decode(&pool)
						if err != nil {
							log.Println(err)
						}
						fmt.Println(pool)
					}

				}

				miner:=rand.Intn(len(pool.Miners))
				pool.Winner=pool.Miners[miner]
				fmt.Println(pool.Miners[miner],people[pool.Miners[miner]].Name,"is selected as miner")



				fmt.Println("Sending Transaction to Miner, i.e.",people[pool.Miners[miner]].Name)
				conn, err := net.Dial("tcp", "localhost:"+people[pool.Miners[miner]].MinePort)
				if err != nil {
					//handle error
					fmt.Println(err)
				}
				gobEncoder := gob.NewEncoder(conn)
				err = gobEncoder.Encode(Transaction{ind,amount,j,pool.Miners[miner],tranfee,pool})

				/*chainHead = ExecuteTransaction(&people[0],amount,&people[j],&people[miner],tranfee,chainHead)
				fmt.Println("Latest Blockchain and Ledger!")
				ListBlocks(chainHead)
				fmt.Println(people)
				fmt.Println("BroadCasting!")
				broadCastBlockChainAndLedger(chainHead,people,0)
				*/
			}else if i ==2{
				for i=0;i<1;i--{

				}
			}
		}

	}else if len(os.Args)==6 {//Peers
		//arg[1]=Name, arg[2]=Port, arg[3]=MinePort, arg[4]=PoolPort, arg[5]=SenderPort
		fmt.Println("This is a new member of the block, named ",os.Args[1])
		conn, err := net.Dial("tcp", "localhost:6000")
		if err != nil {
			//handle error
		}
		person:=&Person{os.Args[1],100,os.Args[2],os.Args[3],os.Args[4],os.Args[5]}
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
		poolln, err := net.Listen("tcp",":"+os.Args[5])
		if err != nil {
			log.Fatal(err)
		}
		go ReceiveUpdatedBlockChain(ln,chainHead,people)
		go ReceivePrizePool(ln,os.Args[4],ind,people)
		go MinerHandler(ln,chainHead,people,os.Args[3],ind)
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

				//New algo
				var pool PrizePool
				pool.Prize=0.0
				fmt.Println("Your transaction is pending to be made part of the blockchain! A prize pool with Minimum and Maximum values is going to be set!")
				pool.MinEntry= float64(rand.Intn(51))
				pool.MaxEntry=pool.MinEntry*2
				pool.SenderPort=people[ind].SenderPort
				//Broadcast Pool
				//BroadCastPrizePool(pool,people,ind)
				//
				fmt.Println("Minimum Entry for the pool is",pool.MinEntry,"And the Maximum Entry is",pool.MaxEntry)
				var entry float64
				fmt.Println("How much do you want to give in the pool?")
				_, err = fmt.Scanln( &entry)
				if err!=nil{
					fmt.Println(err)
				}
				if entry>=pool.MinEntry{
					if entry<=pool.MaxEntry{
						if entry<=people[ind].Wallet{
							people[ind].Wallet-=entry
							pool.Prize+=entry
							pool.Entries = append(pool.Entries, entry)
							pool.Miners = append(pool.Miners, ind)
							//Broadcast Pool

						}else{
							fmt.Println("You donot have",entry,"in your wallet")
						}
					}else{
						fmt.Println("Invalid!")
					}
				}else{
					fmt.Println("Invalid!")
				}
				//pool.Miners
				//Recieve the final prize pool
				for i=0;i<len(people);i++{
					if i!=ind{
						BroadCastPrizePool(pool,people,ind,i)
						fmt.Println("Waiting for the updated pool")
						conn, err := poolln.Accept()
						if err != nil {
							log.Println(err)
						}
						gobEncoder := gob.NewDecoder(conn)
						err = gobEncoder.Decode(&pool)
						if err != nil {
							log.Println(err)
						}
						fmt.Println(pool)
					}

				}

				miner:=rand.Intn(len(pool.Miners))
				pool.Winner=pool.Miners[miner]
				fmt.Println(miner,people[pool.Miners[miner]].Name,"is selected as miner")

				fmt.Println("Sending Transaction to Miner, i.e.",people[pool.Miners[miner]].Name)
				conn, err := net.Dial("tcp", "localhost:"+people[pool.Miners[miner]].MinePort)
				if err != nil {
					//handle error
					fmt.Println(err)
				}
				gobEncoder := gob.NewEncoder(conn)
				err = gobEncoder.Encode(Transaction{ind,amount,j,miner,tranfee,pool})

				/*chainHead = ExecuteTransaction(&people[ind],amount,&people[j],&people[miner],tranfee,chainHead)
				fmt.Println("Latest Blockchain and Ledger!")
				ListBlocks(chainHead)
				fmt.Println(people)
				fmt.Println("BroadCasting!")
				broadCastBlockChainAndLedger(chainHead,people,ind)
				*/

			}else if i ==2{
				for i=0;i<1;i--{

				}
			}
		}

	}

}

