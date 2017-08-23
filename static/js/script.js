$(document).ready(function(){    
    var ipaddr = "128.199.227.5"
    
    var sock = null;
    var wsuri = "ws://" + ipaddr  + "/ws"
    sock = new WebSocket(wsuri);

    var sockd = null;
    var wsurid = "ws://" + ipaddr  + "/wsd"
    sockd = new WebSocket(wsurid);

    sock.onopen = function() {
	//sock.send("ping");
        console.log("connected to " + wsuri);
    }
    sock.onclose = function(e) {
        console.log("connection data closed (" + e.code + ")");
    }
    sock.onmessage = function(e) {
        console.log(e.data)
        var msg = JSON.parse(e.data)
        updateData(msg);
    }
    
    sockd.onopen = function() {
	//sockd.send("ping");
        console.log("connected to " + wsurid);
    }
    sockd.onclose = function(e) {
        console.log("connection debug closed (" + e.code + ")");
    }
    sockd.onmessage = function(e) {       
        console.log(e.data);
        var msg = JSON.parse(e.data);
        updateDebugData(msg);
        notify("Debug", msg.Message);
    }

    var table = document.getElementById("solardata-table");
    function updateData(data) {
        var row = table.insertRow(1);
        var c = new Date(data.created);
        var month = c.getMonth() + 1;
        if(month < 10) {
            month = "0" + month;
        }
        var date = c.getDate();
        if(date < 10) {
            date = "0" + date;
        }
        var hour = c.getHours();
        if (hour < 10) {
            hour = "0" + hour;
        }
        var min = c.getMinutes();
        if (min < 10) {
            min = "0" + min;
        }
        var sec = c.getSeconds();
        if(sec < 10) {
            sec = "0" + sec;
        }
        var cCell = row.insertCell(0);
        cCell.innerHTML = c.getFullYear() + "-" + month + "-" + date +" " + hour + ":" + min + ":" + sec;
        
        var voltage = data.voltage;
        var vCell = row.insertCell(1);
        vCell.innerHTML = voltage;
        var current = data.current;
        var curCell = row.insertCell(2);
        curCell.innerHTML = current;

        var temp1 = data.temp1;
        var t1Cell = row.insertCell(3)
        t1Cell.innerHTML = temp1;
        var temp2 = data.temp2;
        var t2Cell = row.insertCell(4);
        t2Cell.innerHTML = temp2;

        var lum1 = data.lum1;
        var l1Cell = row.insertCell(5);
        l1Cell.innerHTML = lum1;
        var lum2 = data.lum2;
        var l2Cell = row.insertCell(6);
        l2Cell.innerHTML = lum2;

        var pwm = data.pwm;
        var pwmCell = row.insertCell(7);
        pwmCell.innerHTML = pwm;
    }
    
    var debugtable = document.getElementById("solardebug-table");
    function updateDebugData(data) {
        var row = debugtable.insertRow(1);
        var c = new Date(data.Created)
        var cCell = row.insertCell(0);
        cCell.innerHTML = c.toLocaleString();
        
        var message = data.Message;
        var mCell = row.insertCell(1);
        mCell.innerHTML = message;
    }
    
    $("#iv-measurement").click(function(){
        $.ajax({
            url: '/measure',
            type: 'POST',
            data: {measure: 1},
            success: function(result){
                console.log("start measuring");
            }
        })
    });
    
    $("#delete").click(function () {
        $("#confirm-overlay").css("display", "block");
        console.log("delete data");
    });

    $("#delete-confirm").click(function () {
        $.ajax({
            url: '/data',
            type: 'DELETE',
            success: function(result){
                console.log(result);
                window.location.reload();
            }
        });
        console.log("detele confirm");
    });

    $("#delete-cancel").click(function () {
        $("#confirm-overlay").css("display", "none");
        console.log("detele cancel");
    });

    $("#debug-list").click(function(){
        if($("#debug").css("display") == "none") {
            $("#debug").css({"display": "block", "flex-grow": 1})
        } else {
            $("#debug").css("display", "none")
        }        
        console.log("show debug");
    })

    function notify(title, content) {
        $("#notification-title").text(title);
        $("#notification-content").text(content);
        $(".notification").css("display", "block");
        setTimeout(function(){
            $(".notification").css("display", "none");
        }, 5000);
        console.log("notify");
    }
});
