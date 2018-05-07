const mysql = require('mysql');
const connection = mysql.createConnection({
    host: 'localhost', 
    user: 'root', 
    password: 'rVPbyVKrBhHul0Z3', 
    database: 'hw2'});
connection.connect((err) => { 
    if (err) throw err; 
        console.log('Connected!');
    });
module.exports = connection;