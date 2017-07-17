$(document).ready(function(){
    var tempChart = new Chart(document.getElementById("tempChart"), {
        type: 'line',
        data: {
            datasets: [{
                label: 'Temperature1',
                data: [],
                fill: false
            },
            {
                label: 'Temperature2',
                data: [],
                fill: false
            }]
        },
        options: {
            responsive: true,
            scales: {
                xAxes: [{
                    type: 'time',
                    time: {
                        unit: 'hour'
                    }
                }],
                yAxes: [{
                    ticks: {
                        beginAtZero:true
                    }
                }]
            },
            layout: {
                padding: {
                    left: 0,
                    right: 0,
                    top: 0,
                    bottom: 15
                }
            }
        }
    });
    var lumChart = new Chart(document.getElementById("lumChart"), {
        type: 'line',
        data: {
            datasets: [{
                label: 'Luminance1',
                data: [],
                fill: false
            },
            {
                label: 'Luminance2',
                data: [],
                fill: false
            }]
        },
        options: {
            responsive: true,
            scales: {
                xAxes: [{
                    type: 'time',
                    time: {
                        unit: 'hour'
                    }
                }],
                yAxes: [{
                    ticks: {
                        beginAtZero:true
                    }
                }]
            },
            layout: {
                padding: {
                    left: 0,
                    right: 0,
                    top: 0,
                    bottom: 15
                }
            }
        }
    });
    var voltageChart = new Chart(document.getElementById("voltageChart"), {
        type: 'line',
        data: {
            datasets: [{
                label: 'Voltage',
                data: [],
                fill: false
            },
            {
                label: 'Current',
                data: [],
                fill: false
            }]
        },
        options: {
            responsive: true,
            scales: {
                xAxes: [{
                    type: 'time',
                    time: {
                        unit: 'hour'
                    }
                }],
                yAxes: [{
                    ticks: {
                        beginAtZero:true
                    }
                }]
            },
            layout: {
                padding: {
                    left: 0,
                    right: 0,
                    top: 0,
                    bottom: 15
                }
            }
        }
    });

    var table = document.getElementById("solardata-table");

    function updateData(datas) {
        for (var data in datas) {
            var d = data;
            if (typeof datas[data] == 'object') {
                data = datas[data]
            } else {
                data = datas;
            }
            console.log(data);

            var row = table.insertRow(1);

            var c = new Date(data.created)
            var cCell = row.insertCell(0);
            cCell.innerHTML = c.getDate() + "/" + c.getUTCMonth() + "/" + c.getFullYear() + 
                " " + c.getHours() + ":" + c.getMinutes();
            
            var voltage = data.voltage;
            voltageChart.data.datasets[0].data.push({
                    x: c,
                    y: voltage
            });
            var vCell = row.insertCell(1);
            vCell.innerHTML = voltage;

            var current = data.current;
            voltageChart.data.datasets[1].data.push({
                    x: c,
                    y: current
            });
            voltageChart.update();
            var curCell = row.insertCell(1);
            curCell.innerHTML = current;

            var temp1 = data.temp1;
            tempChart.data.datasets[0].data.push({
                    x: c,
                    y: temp1
            });
            var t1Cell = row.insertCell(2)
            t1Cell.innerHTML = temp1;

            var temp2 = data.temp2;
            tempChart.data.datasets[1].data.push({
                    x: c,
                    y: temp2
            });    
            tempChart.update();
            var t2Cell = row.insertCell(3);
            t2Cell.innerHTML = temp2;

            var lum1 = data.lum1;
            lumChart.data.datasets[0].data.push({
                    x: c,
                    y: lum1
            });
            var l1Cell = row.insertCell(4);
            l1Cell.innerHTML = lum1;
            
            var lum2 = data.lum2;
            lumChart.data.datasets[1].data.push({
                    x: c,
                    y: lum2
            });
            lumChart.update();
            var l2Cell = row.insertCell(5);
            l2Cell.innerHTML = lum2;

            if (typeof datas[d] != 'object') {
                return
            } 
        }
    }
    var sock = null;
    var wsuri = "ws://127.0.0.1:8000/ws"

    sock = new WebSocket(wsuri);
    sock.onopen = function() {
        console.log("connected to " + wsuri);
    }

    sock.onclose = function(e) {
        console.log("connection closed (" + e.code + ")");
    }

    sock.onmessage = function(e) {        
        var msg = JSON.parse(e.data)
        updateData(msg);
    }

    $.getJSON('data', function(data) {
        updateData(data);
    })

});

