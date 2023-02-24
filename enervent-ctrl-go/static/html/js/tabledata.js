function zeroPad(number) {
    return ("0" + number).slice(-2)
}

function timeStamp() {
    now = new Date()
    Y = now.getFullYear()
    m = now.getMonth()
    d = now.getDate()
    H = zeroPad(now.getHours())
    M = zeroPad(now.getMinutes())
    S = zeroPad(now.getSeconds())
    return `${Y}-${m}-${d} ${H}:${M}:${S}`
}

function coils(data) {
    if (document.getElementById("coilval_0") == null) {
		for (n=0; n<data.length; n++) {
			tablerow = document.createElement("tr")
			fields = ["address", "value", "symbol", "description"]

			for (i=0; i<fields.length; i++) {
				td = document.createElement("td")
				if (fields[i] == "value") {
					value = document.createTextNode(Number(data[n][fields[i]]))
					td.id = "coilval_" + n;
				} else {
					value = document.createTextNode(data[n][fields[i]])
				}
				td.appendChild(value)
				tablerow.appendChild(td)
			}
			if (data[n].reserved) {
				tablerow.className = "reserved"
				if (!document.getElementById("incl_res").checked) {
					tablerow.hidden = true
				}
			}
			datatable.appendChild(tablerow)
		}
	} else {
		for (n=0; n<data.length; n++) {
			coilval = document.getElementById("coilval_" + n);
			oldval = coilval.innerHTML
			coilval.innerHTML = Number(data[n]["value"])
			if (oldval != coilval.innerHTML) {
				coilval.className = "highlightrow"
				// setTimeout(() => {coilval.className = ""}, 1000)
			} else {
				coilval.className = ""
			}
		}
	}
}

function registers(data) {
	if (document.getElementById("regval_0") == null) {
		console.log(`${timeStamp()} Filling register data...`)
		for (n=0; n<data.length; n++) {
			tablerow = document.createElement("tr")
			fields = ["address", "value", "symbol", "description"]

			for (i=0; i<fields.length; i++) {
				td = document.createElement("td")
				if (fields[i] == "value" && data[n].type == "bitfield") {
					value = document.createTextNode(data[n].bitfield)
				} else {
					value = document.createTextNode(data[n][fields[i]])
				}
				if (fields[i] == "value") {
					td.id = "regval_" + n;
				}
				td.appendChild(value)
				tablerow.appendChild(td)
			}
			if (data[n].reserved) {
				tablerow.className = "reserved"
				if (!document.getElementById("incl_res").checked) {
					tablerow.hidden = true
				}
			}
			datatable.appendChild(tablerow)
		}
		console.log(`${timeStamp()} Done.`)
	} else {
		for (n=0; n<data.length; n++) {
			regval = document.getElementById("regval_" + n);
			oldval = regval.innerHTML
			if (data[n].type == "bitfield") {
				regval.innerHTML = data[n]["bitfield"]
			} else {
				regval.innerHTML = data[n]["value"]
			}
			if (oldval != regval.innerHTML) {
				regval.className = "highlightrow"
				// setTimeout(() => {regval.className = ""}, 1000)
			} else {
				regval.className = ""
			}
		}
	}	
}

function getData() {
    // document.getElementById('time').innerHTML = `${Y}-${m}-${d} ${H}:${M}:${S}`
    document.getElementById('time').innerHTML = timeStamp()
    
    error = false
    // The same index.html is used for both coil and register data,
    // change api url based on which we're looking at
    if (document.location.pathname == "/coils/") {
        url = "/api/v1/coils"
        document.getElementById("title").innerHTML = "Coils | Enervent Pingvin Kotilämpö"
        document.getElementById('caption').innerHTML = "Coil values at "
    }
    else if (document.location.pathname == "/registers/") {
        url = "/api/v1/registers"
        document.getElementById("title").innerHTML = "Registers | Enervent Pingvin Kotilämpö"
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
            if (url == '/api/v1/coils') {
                coils(data)
            } else if (url == '/api/v1/registers') {
                registers(data)
            }
        });
    }

    // Using setTimeout instead of setInterval to avoid possible connection issues
    // There's no need to update exactly every 5 seconds, the skew is fine
    setTimeout(getData, 2*1000);
}

// Show or hide rows for "reserved" values when clicking the checkbox
incl_res = document.getElementById("incl_res")
incl_res.addEventListener("change", (event) => {
	reservedRows = document.getElementsByClassName("reserved")
	if (!event.currentTarget.checked) {
		for (i=0; i<reservedRows.length; i++) {
			reservedRows[i].hidden = true
		}
	} else {
		for (i=0; i<reservedRows.length; i++) {
			reservedRows[i].hidden = false
		}
	}
});

