package main

import (
	_ "embed"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

//go:embed index.html
var page string

type Node struct {
	Path     string
	Children map[string]*Node
	IsWord   bool
}

var dict *Node

func main() {

	log.Println("Starting...")

	var port int
	var dictPath string

	flag.IntVar(&port, "p", 8080, "Port number")
	flag.StringVar(&dictPath, "dict", "", "Path to dictionary")
	flag.Parse()

	if dictPath == "" {
		log.Fatal("Path to dictionary not specified")
	}

	loadDict(dictPath)

	http.HandleFunc("/", handle)
	log.Printf("Starting web server on port %d\n", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))

}

func handle(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	data := struct {
		Letters string
		Words   []string
	}{}

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	data.Letters = r.FormValue("letters")
	data.Words = findAnagrams(data.Letters)

	t, err := template.New("page").Parse(page)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}

}

func addWordToDict(node *Node, head string, tail string) {

	runes := []rune(tail)
	length := len(runes)

	if length > 0 {
		nextLetter := runes[0]
		newPath := head + string(nextLetter)
		nextNode, found := node.Children[newPath]
		if !found {
			nextNode = &Node{
				Path:     newPath,
				Children: make(map[string]*Node),
			}
			node.Children[newPath] = nextNode
		}
		if length == 1 {
			nextNode.IsWord = true
		}
		newTail := string(runes[1:])
		addWordToDict(nextNode, newPath, newTail)
	}

}

func loadDict(path string) {

	log.Printf("Loading dictionary from %s", path)
	dict = &Node{
		Children: make(map[string]*Node),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	strData := strings.ReplaceAll(string(data), "\r", "")
	words := strings.Split(strData, "\n")
	for _, word := range words {
		addWordToDict(dict, "", word)
	}

	log.Printf("Successfully loaded %d words", len(words))

	//printDict(dict)

}

func printDict(node *Node) {
	log.Println(node.Path, node.IsWord)
	for _, child := range node.Children {
		printDict(child)
	}
}

func generateWords(head string, letters []rune, currentNode *Node) []string {

	result := make([]string, 0)

	for i, letter := range letters {
		word := head + string(letter)
		nextNode, found := currentNode.Children[word]
		if found {
			unused := make([]rune, len(letters))
			copy(unused, letters)
			unused = append(unused[:i], unused[i+1:]...)
			if nextNode.IsWord {
				result = append(result, nextNode.Path)
			}
			newWords := generateWords(word, unused, nextNode)
			result = append(result, newWords...)
		}
	}

	return result

}

func findAnagrams(letters string) []string {

	result := generateWords("", []rune(letters), dict)
	mapUnique := make(map[string]struct{})
	unique := make([]string, 0, len(result))

	for _, word := range result {
		mapUnique[word] = struct{}{}
	}

	for k := range mapUnique {
		unique = append(unique, k)
	}

	sort.Strings(unique)

	log.Printf("Anagrams for %s: %v", letters, unique)
	return unique

}
