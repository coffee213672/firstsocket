var app = require('express')();
var http = require('http').Server(app);
var io = require('socket.io')(http);
var userlist = new Array();
var uuserlist = new Array();
var db = require('./conn');
var arrayallsocket = [];


app.get('/', function (req, res) {
  res.sendFile(__dirname + '/index.html');
});

app.get('/:id', function (req, res, next) {
  res.sendFile(__dirname + '/private.html');
});

var roomUser = {};
var rroomUser = {};
io.on('connection', function (socket) {
  var url = socket.request.headers.referer;
  var split_arr = url.split('/');
  var roomid = split_arr[split_arr.length - 1] || 'index';
  var user = '';
  socket.on('login', function (data) {
    user = data;
    if(roomid.length>10){
      roomid = roomid.substring(0,10);
    }
    if (!roomUser[roomid]) {
      roomUser[roomid] = [];
    }
    if (data.name !== null) {
      roomUser[roomid].push({
        'name': data.name,
        'sid': socket.id
      });
      socket.join(roomid);
      var msg = user.name + '加入了房間';
      socket.to(roomid).emit('chat message', msg);
      socket.emit('chat message', msg);
      var oon = new Array();
      var ioon = new Array();
      if (roomid !== 'index') {
        oon = roomUser[roomid].map(function (o) {
          return o.name;
        });
        socket.to(roomid).emit('userlist', oon);
        socket.emit('userlist', oon);
      } else {
        ioon = roomUser['index'].map(function (o) { 
          return o.name;
        });
        socket.to('index').emit('userlist', ioon);
        socket.emit('userlist', ioon);
        db.query("SELECT * FROM ( SELECT * FROM c_message ORDER BY id DESC LIMIT 10 ) sub ORDER BY id ASC", (err, rows) => {
          if (err) {
            console.log("no");
          } else {
            var xh = new Array();
            for (var i = 0; i < rows.length; i++) {
              var hstr = '';
              xh = rows[i];
              hstr = '' + rows[i]['nickname'] + ' : ' + rows[i]['message'] + "  (" + GMTToStr(rows[i]['time']) + ")";
              socket.emit('chat message', hstr);
            }
          }
        })
      }
    } else {
      socket.disconnect();
    }
  });

  socket.on('chat message', function (msg) {
    var ndate = new Date();
    var r = GetNowDate(ndate);
    msg = msg.trim();
    if (msg.substr(0, 3) !== '/w ') {
      if (roomUser[roomid].map(function (e) {
          return e.name
        }).indexOf(user.name) < 0) {
        return false;
      }
      var amsg = user.name + '：' + msg + ' ' + '(' + r + ')';
      socket.to(roomid).emit('chat message', amsg);
      socket.emit('chat message', amsg);
      if (roomid == 'index') {
        db.query("insert into c_message(nickname,message,time) values('" + user.name + "','" + msg + "','" + r + "')", (err, rows) => {
          if (err) throw err;

        });
      }
    } else {
      p_msg = msg.substr(3);
      var name_loc = p_msg.indexOf(' ');
      var fname = p_msg.substring(0, name_loc);
      var fmsg = p_msg.substring(name_loc + 1);
      //if (fname in roomUser['index']) {
      var t = roomUser['index'].filter(function (item) {
        return item.name == fname;
      })
      var tt = t.map(t => t.sid);
      var ttt = tt.toString();
      var d = roomUser['index'].filter(function (item) {
        return item.name == user.name;
      })
      var dd = d.map(d => d.sid);
      var ddd = dd.toString();
      if (fname !== user.name) {
        var pmsg = '來自' + user.name + '給你的訊息 : ' + fmsg + "  (" + r + ")";
        var ppmsg = '發給' + fname + '的訊息 : ' + fmsg + "  (" + r + ")";
        io.sockets.sockets[ddd].emit('chat message', ppmsg);
        io.sockets.sockets[ttt].emit('chat message', pmsg);
      } else {
        var pmsg = '切勿自言自語';
        io.sockets.sockets[ttt].emit('chat message', pmsg);
      }
    }
  });


  socket.on('private room', (iname, room_name) => {

    if (typeof roomUser[room_name] === 'undefined') {
      roomUser[room_name] = []
    }
    var t = roomUser['index'].filter(function (item) {
      return item.name == user.name;
    })
    var tt = t.map(t => t.sid);
    var ttt = tt.toString();

    var q = roomUser['index'].filter(function (item) {
      return item.name == iname;
    })
    var qq = q.map(q => q.sid);
    var qqq = qq.toString();

    var purl = 'http://localhost:3000/'+room_name+'?'+user.name;
    var bpurl = 'http://localhost:3000/'+room_name+'?'+iname;
    io.sockets.sockets[ttt].emit('private url',purl);
    io.sockets.sockets[qqq].emit('private url',bpurl);
    //測試喔= =
  })



  socket.on('kick', function (data) {
    if (roomid === 'index') {
      var k = roomUser[roomid].filter(function (item) {
        return item.name == data;
      })
      var kk = k.map(k => k.sid);
      var kkk = kk.toString();
      var index = roomUser[roomid].map(function (e) {
        return e.sid
      }).indexOf(kkk);
      if (kkk !== '') {
        var ks = 'kick';
        io.sockets.sockets[kkk].disconnect(ks);
      }
      roomUser[roomid].splice(index, 1);
      var lon = roomUser[roomid].map(function (o) {
        return o.name;
      })
      var kmsg = data + ' 被 ' + user.name + ' 踢除大廳';
      socket.to(roomid).emit('userlist', lon);
      socket.emit('userlist', lon);
      socket.to(roomid).emit('chat message', kmsg);
      socket.emit('chat message', kmsg);
    }
  })

  socket.on('disconnect', function (ks) {
    socket.leave(roomid, function (err) {
      if (err) {
        log.error(err);
      } else if (ks !== 'server namespace disconnect' && typeof roomUser[roomid] !== 'undefined') {
        var k = roomUser[roomid].filter(function (item) {
          return item.name == user.name;
        })

        var kk = k.map(k => k.sid);
        var kkk = kk.toString();
        var index = roomUser[roomid].map(function (e) {
          return e.sid
        }).indexOf(kkk);
        console.log(index);
        if (index !== -1) {
          roomUser[roomid].splice(index, 1);
          if (user.name !== 'undefined' && index !== -1) {
            var lmsg = user.name + '退出了房間';
            var lon = roomUser[roomid].map(function (o) {
              return o.name;
            })
            socket.to(roomid).emit('chat message', lmsg);
          }
          socket.to(roomid).emit('userlist', lon);
          socket.emit('userlist', lon);
        }
      }
    });
  });
});


function GetNowDate(type) {
  var date = new Date();
  var year = date.getFullYear();

  var month = (1 + date.getMonth()).toString();
  month = month.length > 1 ? month : '0' + month;

  var day = date.getDate().toString();
  day = day.length > 1 ? day : '0' + day;

  var Hour = date.getHours().toString();
  Hour = Hour.length > 1 ? Hour : '0' + Hour;

  var Minute = date.getMinutes().toString();
  Minute = Minute.length > 1 ? Minute : '0' + Minute;

  var Second = date.getSeconds().toString();
  Second = Second.length > 1 ? Second : '0' + Second;
  if (type == 0)
    return Hour + ":" + Minute + ":" + Second;
  else
    return year + "-" + month + "-" + day + " " + Hour + ":" + Minute + ":" + Second;
  //return year + month +  day + Hour + Minute + Second;
}

function RandomSn(len) {
  var keylist = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUV123456789"
  temp = ''
  for (i = 0; i < len; i++)
    temp += keylist.charAt(Math.floor(Math.random() * keylist.length))
  return temp
}

function GMTToStr(time) {
  let date = new Date(time)
  let Str = date.getFullYear() + '-' +
    (date.getMonth() + 1) + '-' +
    date.getDate() + ' ' +
    date.getHours() + ':' +
    date.getMinutes() + ':' +
    date.getSeconds()
  return Str
}

http.listen(3000, function () {
  console.log('listening on *:3000');
});