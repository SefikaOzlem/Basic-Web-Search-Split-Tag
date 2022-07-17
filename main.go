package main

import (
	"database/sql"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"

	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	//	"time"
)

var tmpl = template.Must(template.ParseFiles("edit.html"))
var tpl = template.Must(template.ParseFiles("index.html"))
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY"))) // for using get session ıd
var exist_count int
var searchQuery string // temp to search input
var count int

// temp to number of image url in web page
var temp_imgsrc string
var temp_userIp string
var temp_sessionId string

var user_name string
var pass_word string

var search_count int

type DataAddress struct {
	ip_address string // user's ip address
	session_id string // session id
}

var db, errr = sql.Open("sqlite3", "db.db")
var flag bool

var err error

var databaseUsername string
var databasePassword string

var time_date string

var pros, eror = db.Prepare("Insert into panel_info_users (user_session,user_ip,enter_time) values (?,?,?)")
var search_info, errors = db.Prepare("Insert into panel_info_search (user,user_ip,search_site,time) values (?,?,?,?)")

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,"<html>\n<head>\n<style>\nbody {\n  background-image: url('https://images.pexels.com/photos/19670/pexels-photo.jpg?auto=compress&cs=tinysrgb&dpr=2&h=650&w=940');\n  background-repeat: no-repeat;\n  background-attachment: fixed; \n  background-size: 100% 100%;\n}\n</style>\n</head>\n<body></body>\n</html>")

	fmt.Println(time_date)
	dizin, err := os.Open("assets")
	if err != nil {
		os.Exit(1)
	}
	defer dizin.Close()
	liste, _ := dizin.Readdirnames(0)
	count = len(liste)

	// Find ip address
	url := "https://api.ipify.org?format=text"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Find session id
	cookie, err := r.Cookie("session")
	if err != nil {
		id, _ := uuid.NewRandom()
		cookie = &http.Cookie{
			Name:  "session",
			Value: id.String(),
			// Secure: true,
			HttpOnly: true,
			Path:     "/",
		}
		http.SetCookie(w, cookie)
	}

	datas := DataAddress{
		ip_address: string(ip),
		session_id: cookie.Value,
	}
	temp_userIp = datas.ip_address
	temp_sessionId = datas.session_id

	exist, _ := db.Query("SELECT user_session FROM panel_info_users")
	var user_session string

	for exist.Next() { // Iterate and fetch the records from result cursor
		exist.Scan(&user_session)
		if strings.Compare(user_session, temp_sessionId) == 0 {
			exist_count++
			fmt.Println(user_session)
			fmt.Println("Kullanıcı  zaten mevcut")
		}
	}
	fmt.Println("exist _count:", exist_count)
	if exist_count == 0 {
		currentTime := time.Now()
		time_date = currentTime.Format("02.01.2006 15:04:05")
		data, _ := pros.Exec(temp_sessionId, temp_userIp, time_date)
		print(data)
	}
	exist_count = 0

	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		fmt.Fprintf(w, "<html> <body><p style=\"text-align:center\"><font size=5 color=#00008b>Kullanıcı IP Address:</color></font><font size=4color=#696969 >"+datas.ip_address+"</br> </color></font><font size=5 color=#00008b>Session Id:</color></font><font size=4color=#696969 >"+datas.session_id+" </color></font></p></body></html>")
		return
	}
	tmpl.Execute(w, struct{ Success bool }{true})

}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	var pageTitle string       // title of the web page ( in html )
	var metaDescription string // description of the web page (in html ... <meta)
	var keywords string        // keywords of the web page (in html ... <meta)

	u, err := url.Parse(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	params := u.Query()
	searchQuery = params.Get("q") // search input
	page := params.Get("page")

	currentTime := time.Now()
	time_date = currentTime.Format("02.01.2006 15:04:05")
	data, _ := search_info.Exec(temp_sessionId, temp_userIp, searchQuery, time_date)
	print(data)

	if page == "" {
		page = "1"
	}
	search_count++
	var urlProtocolSheme = "http://"
	//Find domain address
	url_temp, error := url.Parse(urlProtocolSheme + searchQuery)
	if error != nil {
		log.Fatal(error)
	}
	parts := strings.Split(url_temp.Hostname(), ".")
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]

	fmt.Println(domain)

	doc, err := goquery.NewDocument(urlProtocolSheme + searchQuery)
	if page == "1" {
		fmt.Println(urlProtocolSheme + searchQuery)
		if err != nil {
			log.Fatal(err)
		}

		pageTitle = doc.Find("title").Contents().Text()
		doc.Find("meta").Each(func(index int, item *goquery.Selection) {

			if item.AttrOr("name", "") == "description" {
				metaDescription = item.AttrOr("content", "")
			}

			if metaDescription == "" {
				if item.AttrOr("property", "") == "og:description" {
					metaDescription = item.AttrOr("content", "")
				} else if item.AttrOr("name", "") == "Description" {
					metaDescription = item.AttrOr("content", "")
				}
			}

			if item.AttrOr("name", "") == "keywords" {
				keywords = item.AttrOr("content", "")
			}

			if keywords == "" {
				if item.AttrOr("property", "") == "og:keywords" {
					keywords = item.AttrOr("content", "")
				} else if item.AttrOr("name", "") == "Keywords" {
					keywords = item.AttrOr("content", "")
				}
			}

			if pageTitle == "" {
				if item.AttrOr("property", "") == "og:title" {
					pageTitle = item.AttrOr("content", "")
				}
			}

		})
		fmt.Fprintf(w, "<html> <body> <div class='group'> <br/> "+"Title:"+pageTitle+"</div><div class='group'> <br/> "+"Description:"+metaDescription+"</div><div class='group'> <br/> "+"Keywords:"+keywords+"</div><br/></body></html>")

		// Find and print image URLs
		doc.Find("img ").Each(func(index int, element *goquery.Selection) {
			imgSrc, exists := element.Attr("src")
			if exists {
				if imgSrc != "" {
					count++
					if strings.HasPrefix(imgSrc, "http") == false {
						temp_imgsrc = imgSrc
						imgSrc = "http://" + domain + "/" + imgSrc

					}
				}

				if imgSrc != "" && (strings.HasSuffix(imgSrc, ".png") == true || strings.HasSuffix(imgSrc, ".jpeg") == true || strings.HasSuffix(imgSrc, ".jpg") == true) {
					fmt.Println("Image url: ", imgSrc)
					url := imgSrc
					// don't worry about errors
					response, e := http.Get(url)
					if e != nil {
						url = "http://" + searchQuery + "/" + temp_imgsrc
						fmt.Println(url)
						//log.Fatal(e)

					}
					response, e = http.Get(url)
					defer response.Body.Close()
					img_file := "img" + strconv.Itoa(count) + ".jpg"
					//open a file for writing
					file, err := os.Create("assets/" + img_file)
					if err != nil {
						log.Fatal(err)
					}
					defer file.Close()

					// Use io.Copy to just dump the response body to the file. This supports huge files
					_, err = io.Copy(file, response.Body)
					if err != nil {
						log.Fatal(err)
					}

					fileName := "img" + strconv.Itoa(count) + ".jpg"
					fmt.Println(fileName)
					//fmt.Fprintf(w,"<a style='text-decoration: none; color: orange;'><img src='/assets/"+fileName+"' style='width: 140px'>" +"<div style='width: 130px; text-align: center;'>I just love .</div>\n</a> ")
					//fmt.Fprintf(w, "<html><body><img src='/assets/"+fileName+"'style='margin:5px;width:225px;height:205px';</body></br></html>")
					//	fmt.Fprintf(w,"<html> <body>"+"Image Url:"+strconv.Itoa(count) +"</body></html>")
					if element.AttrOr("alt", "") != "" {
						fmt.Println("alt", element.AttrOr("alt", ""))
						fmt.Fprintf(w, "<html><body><img src='/assets/"+fileName+"'style='margin:5px;width:225px;height:205px;'<p>"+"Image Url:"+imgSrc+"	Alt:"+element.AttrOr("alt", "")+"</p></body></br></html>")

					} else {
						fmt.Fprintf(w, "<html><body><img src='/assets/"+fileName+"'style='margin:5px;width:225px;height:205px;'<p>"+"Image Url:"+imgSrc+"</p></body></br></html>")

					}

				}
			}

		})

	}

	fmt.Println("Search Query is: ", searchQuery)
	fmt.Println("Page is: ", page)

	//keeping log records
	file, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)
	log.Printf("Ip address: %s ; Session id: %s ; Url: %s  ;  Domain; %s \t", temp_userIp, temp_sessionId, urlProtocolSheme+searchQuery, domain)

}

func adminPage(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "<html><body><h1 style='font-family: Cambria Math' ;>Report Of Website </h1></body></html>")

	fmt.Println(time_date)
	fmt.Fprintf(res, "<html>\n<head>\n<style>\nbody {\n  background-image: url('https://image.freepik.com/free-vector/vibrant-pink-watercolor-painting-background_53876-58930.jpg');\n  background-repeat: no-repeat;\n  background-attachment: fixed; \n  background-size: 100% 100%;\n}\n</style>\n</head>\n<body></body>\n</html>")
	dizin, err := os.Open("assets")
	if err != nil {
		os.Exit(1)
	}
	defer dizin.Close()
	liste, _ := dizin.Readdirnames(0)

	rows, _ := db.Query("SELECT COUNT (*) as count FROM panel_info_users")
	var userCount int
	for rows.Next() { // Iterate and fetch the records from result cursor
		rows.Scan(&userCount)
	}
	fmt.Println(userCount)

	rows, _ = db.Query("SELECT COUNT (*) as count FROM panel_info_search")
	var searchCount int
	for rows.Next() { // Iterate and fetch the records from result cursor
		rows.Scan(&searchCount)
	}
	fmt.Println(searchCount)

	fmt.Fprintf(res, "<html>\n<head>\n<style>\n.button {\n  background-color: #4CAF50; /* Green */\n  border: none;\n  color: white;\n  padding: 16px 32px;\n  text-align: center;\n  text-decoration: none;\n  display: inline-block;\n  font-size: 16px;\n  margin: 30px 50px;\n  transition-duration: 0.4s;\n  cursor: pointer;\n}\n\n.button1 {\n  background-color: white; \n  color: black; \n  border: 2px solid #4CAF50;\n}\n\n.button1:hover {\n  background-color: #4CAF50;\n  color: white;\n}\n\n.button2 {\n  background-color: white; \n  color: black; \n  border: 2px solid #008CBA;\n}\n\n.button2:hover {\n  background-color: #008CBA;\n  color: white;\n}\n\n.button3 {\n  background-color: white; \n  color: black; \n  border: 2px solid #f44336;\n}\n\n.button3:hover {\n  background-color: #f44336;\n  color: white;\n}  </style>\n</head>\n<body><button class=\"button button1\">Number of Users: "+strconv.Itoa(userCount)+"</button>\n<button class=\"button button2\">Number of Search: "+strconv.Itoa(searchCount)+"</button>\n<button class=\"button button3\">Number of images: "+strconv.Itoa(len(liste))+"</button>\n</body>\n</html>")

	fmt.Fprintf(res, "<html><body><h1 style='font-family: Cambria Math';>Users Information</h1></body></html>")
	rows, _ = db.Query("SELECT * FROM panel_info_users")
	var userSession string
	var userIp string
	var time string
	fmt.Fprintf(res, "<html><head><style> table, th, td {\n  border: 1px solid black;\n  border-collapse: collapse;\n}\nth, td \n</style></head><body><table style='width:700'><tr><th style='width:300' ;>User Session </th> <th style='width:100' ;>User Ip</th><th style='width:150' ;>Enter Time</th></tr></table></body></html>")

	for rows.Next() { // Iterate and fetch the records from result cursor
		rows.Scan(&userSession, &userIp, &time)
		fmt.Fprintf(res, "<html><head><style> table, th, td {\n  border: 1px solid black;\n  border-collapse: collapse;\n}\nth, td \n</style></head><body><table style='width:700'><tr><td style='width:300'>"+userSession+"</td><td style='width:100'>"+userIp+"</td><td style='width:150'>"+time+"</td></tr> </table></body></html>")

	}

	rows, _ = db.Query("SELECT * FROM panel_info_search")
	var user string
	var user_ip string
	var search_site string
	var time_info string
	fmt.Fprintf(res, "<html><body><h1 style='font-family: Cambria Math';>Search Information </h1></body></html>")
	fmt.Fprintf(res, "<html><head><style> table, th, td {\n  border: 1px solid black;\n  border-collapse: collapse;\n}\nth, td \n</style></head><body><table style='width:900'><tr><th style='width:300' ;>User Session </th> <th style='width:110' ;>User Ip</th><th style='width:340' ;>Search Site </th><th style='width:150' ;>Time</th></tr></table></body></html>")

	for rows.Next() { // Iterate and fetch the records from result cursor
		rows.Scan(&user, &user_ip, &search_site, &time_info)
		fmt.Fprintf(res, "<html><head><style> table, th, td {\n  border: 1px solid black;\n  border-collapse: collapse;\n}\nth, td \n</style></head><body><table style='width:900'><tr><td style='width:300'>"+user+"</td><td style='width:110'>"+user_ip+"</td><td style='width:340'>"+search_site+"</td><td style='width:150'>"+time_info+"</td></tr> </table></body></html>")

	}
	fmt.Fprintf(res, "<html><body><h1 style='font-family: Cambria Math';>Images </h1></body></html>")

	for _, isim := range liste {
		fmt.Println(isim)
		fmt.Fprintf(res, "<html><body><img src='/assets/"+isim+"' style='margin:5px;width:225px;height:205px;'</body></html>")

	}

}

func main() {

	if errr != nil {
		panic(errr.Error())
	}
	defer db.Close()
	rootdir, err := os.Getwd()
	if err != nil {
		rootdir = "No dice"
	}

	//mux := http.NewServeMux()
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/panel", adminPage)
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir(path.Join(rootdir, "assets/")))))
	http.ListenAndServe(":8080", nil)
}
