package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

type Arguments map[string]string

//Used to work with user input at -item option
type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

const fileP = 0644

func Perform(args Arguments, writer io.Writer) error {
	if v, ok := args["fileName"]; !ok || v == "" {
		return errors.New("-fileName flag has to be specified")
	}

	if v, ok := args["operation"]; !ok || v == "" {
		fmt.Println(args["opertaion"])
		return errors.New("-operation flag has to be specified")
	}

	switch args["operation"] {
	case "list":
		err := List(args["fileName"], writer)
		return err
	case "add":
		err := Add(args, writer)
		return err
	case "findById":
		err := FindById(args, writer)
		return err
	case "remove":
		err := Remove(args, writer)
		return err
	default:
		return fmt.Errorf("Operation %s not allowed!", args["operation"])
	}
}

func List(fileName string, writer io.Writer) error {

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, fileP)
	if err != nil {
		return fmt.Errorf("failed to opern file %v", fileName)
	}
	defer file.Close()

	res, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file %v", fileName)
	}

	if _, err := writer.Write(res); err != nil {
		return fmt.Errorf("failed to write users to %T", writer)
	}

	return nil
}

func Add(args Arguments, writer io.Writer) error {
	var userInput User
	if args["item"] == "" {
		return errors.New("-item flag has to be specified")
	} else {
		if err := json.Unmarshal([]byte(args["item"]), &userInput); err != nil {
			return fmt.Errorf("failed to Unmarshal -item input from cli: %w", err) 
		}
	}

	file, err := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, fileP)
	if err != nil {
		return fmt.Errorf("failed to open/create %v to add user: %w", args["fileName"], err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to recieve Stat value of %v: %w", args["fileName"], err) 
	}

	var users []User

	//If file just created, need to adjust prfx to the start of the file
	prfx := "["
	if stat.Size() > 0 {

	        bytes, err := io.ReadAll(file)
	        if err != nil {
			return fmt.Errorf("failed to read the contents of %v to check existing of user: %w", args["fileName"], err)
	        }

	        if err := json.Unmarshal(bytes, &users); err != nil {
			return fmt.Errorf("failed to unmarshal json file %v: %w", args["fileName"], err)
	        }

		//Search if user is already exists
	        for _, u := range users {
			if u.Id == userInput.Id {
				if _, err := writer.Write([]byte(fmt.Sprintf("Item with id %v already exists", u.Id))); err != nil {
					return fmt.Errorf("failed to write to %T message about existing user: %w", writer, err)
				}
				return nil
			}
	        }

		//If no user exists, added it to the end of file
		prfx = ",\n"
		if _, err := file.Seek(-1, 2); err != nil {
			return fmt.Errorf("failed to seek in file %v: %w", args["fileName"], err)
		}

	}

	if _, err := file.WriteString(prfx); err != nil {
		return fmt.Errorf("failed to write prefix %v to file %v: %w", prfx, args["fileName"], err)
	}

	enc := json.NewEncoder(file)
	if err := enc.Encode(userInput); err != nil {
		return fmt.Errorf("failed to Encode userInput to json file %v: %w", args["fileName"], err)
	}

	if _, err := file.Seek(-1, 2); err != nil {
		return fmt.Errorf("failed to seek in file %v: %w", args["fileName"], err)
	}

	if _, err := file.WriteString("]"); err != nil {
		return fmt.Errorf("failed to write suffix %v to file %v: %w", "]", args["fileName"], err)
	}

	return nil
}

//Simple linear search for the User by ID
func FindById(args Arguments, writer io.Writer) error {
	if args["id"] == "" {
		return errors.New("-id flag has to be specified")
	}

	file, err := os.Open(args["fileName"])
	if err != nil {
		return fmt.Errorf("failed to open %v for reading: %w", args["fileName"], err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read the contents of %v to find user: %w", args["fileName"], err)
	}

	var users []User
	if err := json.Unmarshal(bytes, &users); err != nil {
		return fmt.Errorf("failed to unmarshal json file %v: %w", args["fileName"], err)
	}

	//Search for user and output it to writer
	for _, u := range users {
		if u.Id == args["id"] {
			jsn, err := json.Marshal(u)
			if err != nil {
				return fmt.Errorf("failed to marshal user %v in file %v: %w", u, args["fileName"], err)
			}

			if _, err := writer.Write([]byte(jsn)); err != nil {
				return fmt.Errorf("failed to write to %T needed user %v: %w", writer, u, err)
			}
			return nil
		}
	}

	//If we don'f find User with needed ID, just output "" to writer
	if _, err := writer.Write([]byte("")); err != nil {
		return fmt.Errorf("failed to write blank line to %T: %w", writer, err)
	}
	return nil
}

//If finded - remove user from byte slice and then write to truncated (0 size) file
func Remove(args Arguments, writer io.Writer) error {
	if args["id"] == "" {
		return errors.New("-id flag has to be specified")
	}

	file, err := os.OpenFile(args["fileName"], os.O_RDWR, fileP)
	if err != nil {
		return fmt.Errorf("failed to open file %v to remove user: %w", args["fileName"], err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read the contents of %v to remove user: %w", args["fileName"], err)
	}

	var users []User
	if err := json.Unmarshal(bytes, &users); err != nil {
		return fmt.Errorf("failed to unmarshal json file %v: %w", args["fileName"], err)
	}

	for i, u := range users {
		if u.Id == args["id"] {
			users = append(users[:i], users[i+1:]...)

			jsn, err := json.Marshal(users)
			if err != nil {
				return fmt.Errorf("failed to marshal users %v before writing to file %v: %w", u, args["fileName"], err)
			}

			file.Truncate(0)
			file.Seek(0, 0)
			_, err = file.Write([]byte(jsn))
			if err != nil {
				return fmt.Errorf("failed to rewrite file %v after deleting user %v: %w", args["fileName"], u, err)
			}
			return nil
		}
	}

	if _, err := writer.Write([]byte(fmt.Sprintf("Item with id %v not found", args["id"]))); err != nil {
		return fmt.Errorf("failed to output message to writer %T: %w", writer, err)
	}
	return nil
}

func parseArgs() Arguments {
	var oFlag = flag.String("operation", "", "Choose \"add\",\"list\",\"findById\" or \"remove\" operation.")
	var idFlag = flag.String("id", "", "Enter ID \"id\" 1")
	var iFlag = flag.String("item", "", "Enter user `{\"id\": \"1\", \"email\":\"email@test.com\",\"age\": 23}`")
	var fFlag = flag.String("fileName", "", "Enter file \"users.json\"")

	flag.Parse()

	args := Arguments{
		"id":        *idFlag,
		"operation": *oFlag,
		"item":      *iFlag,
		"fileName":  *fFlag,
	}
	return args
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
