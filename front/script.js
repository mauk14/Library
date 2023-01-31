const requestURL = "http://localhost:4000/v1/healthcheck"
let headers = new Headers();

headers.append('Content-Type', 'application/json');
headers.append('Accept', 'application/json');

headers.append('Access-Control-Allow-Origin', 'http://localhost:4000');
headers.append('Access-Control-Allow-Credentials', 'true');

headers.append('GET', 'POST', 'OPTIONS');


const xhr = new XMLHttpRequest()

xhr.open("GET", requestURL)

// xhr.responseType = "json"

// xhr.onload = () => {
//     document.getElementsByClassName("kruto").innerHTML = xhr.response.status
// }

xhr.send()