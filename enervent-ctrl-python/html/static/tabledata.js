// TODO:    JS check
//          checkboxes for live/reserved
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
    
    if (document.location.pathname == "/coils/") {
        url = "/api/v1/coils"
    }
    else if (document.location.pathname == "/registers/") {
        url = "/api/v1/registers"
    }
    error = false
    fetch(url)
    .then((response) => {
        if (!response.ok) {
            throw new Error(`Error fetching data: ${response.status}`)
        }
        return response.json()
    })
    .then((coils) => {
        document.getElementById('coildata').innerHTML = "";
        for (n=0; n<coils.length; n++) {
            tablerow = `<tr><td class="addr" id="addr_${coils[n].address}">${coils[n].address}</td>\
                            <td class ="val" id="value_${coils[n].address}">${Number(coils[n].value)}</td>\
                            <td class="symbol" id="symbol_${coils[n].address}">${coils[n].symbol}</td>\
                            <td class="desc" id="description_${coils[n].address}">${coils[n].description}</td></tr>`
            document.getElementById('coildata').innerHTML += tablerow
        }
    });

    if (!error) setTimeout(getData, 5*1000);
}