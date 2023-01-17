function zeroPad(number) {
    return ("0" + number).slice(-2)
}
function getData() {
    now = new Date()
    Y = now.getFullYear()
    m = now.getMonth()
    d = now.getDate()
    H = zeroPad(now.getHours())
    M = zeroPad(now.getMinutes())
    S = zeroPad(now.getSeconds())
    document.getElementById('time').innerHTML = `${Y}-${m}-${d} ${H}:${M}:${S}`
    
    error = false
    // The same index.html is used for both coil and register data,
    // change api url based on which we're looking at
    if (document.location.pathname == "/coils/") {
        url = "/api/v1/coils"
    }
    else if (document.location.pathname == "/registers/") {
        url = "/api/v1/registers"
    }
    else {
        document.getElementById("data").innerHTML = 'Page not found'
        error = true
    }
    if (!error) {
        // Fetch data from API
        fetch(url)
        .then((response) => {
            if (!response.ok) {
                throw new Error(`Error fetching data: ${response.status}`)
            }
            return response.json()
        })
        .then((data) => {
            // Populate table
            document.getElementById('coildata').innerHTML = "";
            for (n=0; n<data.length; n++) {
                tablerow = `<tr><td class="addr" id="addr_${data[n].address}">${data[n].address}</td>\
                                <td class ="val" id="value_${data[n].address}">${Number(data[n].value)}</td>\
                                <td class="symbol" id="symbol_${data[n].address}">${data[n].symbol}</td>\
                                <td class="desc" id="description_${data[n].address}">${data[n].description}</td></tr>`
                document.getElementById('coildata').innerHTML += tablerow
            }
        });
    }

    // Using setTimeout instead of setInterval to avoid possible connection issues
    // There's no need to update exactly every 5 seconds, the skew is fine
    setTimeout(getData, 5*1000);
}