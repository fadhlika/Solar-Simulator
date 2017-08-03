$(document).ready(function(){
    var tempChart = new Chart(document.getElementById("tempChart"), {
        type: 'line',
        data: {
            datasets: [{
                pointRadius: 0,
                backgroundColor: 'rgb(153, 153, 153)',
                borderColor: 'rgb(153, 153, 153)',
                label: 'Temperature1',
                data: [],
                fill: false
            },
            {
                pointRadius: 0,
                backgroundColor: 'rgb(58, 58, 58)',
                borderColor: 'rgb(58, 58, 58)',
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
            }
        }
    });
    var lumChart = new Chart(document.getElementById("lumChart"), {
        type: 'line',
        data: {
            datasets: [{
                pointRadius: 0,
                backgroundColor: 'rgb(153, 153, 153)',
                borderColor: 'rgb(153, 153, 153)',
                label: 'Luminance1',
                data: [],
                fill: false
            },
            {
                pointRadius: 0,
                backgroundColor: 'rgb(58, 58, 58)',
                borderColor: 'rgb(58, 58, 58)',
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
            }
        }
    });
    var voltageChart = new Chart(document.getElementById("voltageChart"), {
        type: 'line',
        data: {
            datasets: [{
                pointRadius: 0,
                backgroundColor: 'rgb(153, 153, 153)',
                borderColor: 'rgb(153, 153, 153)',
                label: 'Voltage',
                data: [],
                fill: false
            },
            {
                pointRadius: 0,
                backgroundColor: 'rgb(58, 58, 58)',
                borderColor: 'rgb(58, 58, 58)',
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
            }
        }
    });

    function updateData(datas) {
        for (var data in datas) {
            var d = data;
            if (typeof datas[data] == 'object') {
                data = datas[data]
            } else {
                data = datas;
            }


            var c = new Date(data.created)            
            var voltage = data.voltage;
            voltageChart.data.datasets[0].data.push({
                    x: c,
                    y: voltage
            });
            var current = data.current;
            voltageChart.data.datasets[1].data.push({
                    x: c,
                    y: current
            });
            voltageChart.update();
            var temp1 = data.temp1;
            tempChart.data.datasets[0].data.push({
                    x: c,
                    y: temp1
            });
            var temp2 = data.temp2;
            tempChart.data.datasets[1].data.push({
                    x: c,
                    y: temp2
            });    
            tempChart.update();
            var lum1 = data.lum1;
            lumChart.data.datasets[0].data.push({
                    x: c,
                    y: lum1
            });
            var lum2 = data.lum2;
            lumChart.data.datasets[1].data.push({
                    x: c,
                    y: lum2
            });
            lumChart.update();
            if (typeof datas[d] != 'object') {
                return
            } 
        }
    }

    

    var sock = null;
    var wsuri = "ws://128.199.162.40/ws"

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
        showMain();
    })
    
    function showMain() {
        document.getElementById("loader").style.display = "none";
        document.getElementById("main").style.display = "flex";
    }

    var debugtable = document.getElementById("solardebug-table");

    function updateDebugData(datas) {
        for (var data in datas) {
            var d = data;
            if (typeof datas[data] == 'object') {
                data = datas[data]
            } else {
                data = datas;
            }

            var row = debugtable.insertRow(1);

            var c = new Date(data.Created)
            var cCell = row.insertCell(0);
            cCell.innerHTML = c.toLocaleString();
            
            var message = data.Message;
            var mCell = row.insertCell(1);
            mCell.innerHTML = message;

            if (typeof datas[d] != 'object') {
                return
            } 
        }
    }
    var sockd = null;
    var wsurid = "ws://128.199.162.40/wsd"

    sockd = new WebSocket(wsurid);
    sockd.onopen = function() {
        console.log("connected to " + wsurid);
    }

    sockd.onclose = function(e) {
        console.log("connection closed (" + e.code + ")");
    }

    sockd.onmessage = function(e) {        
        var msg = JSON.parse(e.data)
        updateDebugData(msg)
    }

    $.getJSON('debug', function(data) {
        updateDebugData(data);
    })

    function deleteData() {
    document.getElementById("confirm-overlay").style.display = "block";
    }

    function confirmDelete() {
        $.ajax({
            url: '/data',
            type: 'DELETE',
            success: function(result){
                console.log(result);
                window.location.reload();
            }
        });
    }

    function cancelDelete() {
        document.getElementById("confirm-overlay").style.display = "none"
    }
});