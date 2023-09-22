// snippet of code @ 2023-09-17 18:11:52

// === Rust Playground ===
// This snippet is in: /home/yuansl/src/playground/at-2023-09-17-181152/

// Execute the snippet: C-c C-c
// Delete the snippet completely: C-c k
// Toggle between main.rs and Cargo.toml: C-c b

use std::{borrow::Borrow, io};

trait Writer {
    type Size;
    fn write(&self, buf: Vec<u8>) -> Result<Self::Size, io::Error>;
}

struct Some(String);

impl Writer for Some {
    type Size = usize;
    fn write(&self, buf: Vec<u8>) -> Result<Self::Size, io::Error> {
        Ok(buf.len())
    }
}

impl Some {
    fn string(&self) -> String {
        self.0.parse().unwrap()
    }
}

impl Writer for String {
    type Size = usize;
    fn write(&self, _: Vec<u8>) -> Result<Self::Size, io::Error> {
        Ok(self.len())
    }
}

fn foo4() -> String {
    let mut x = [3343, -1, 3434, 4302439, 00039343, 4334];
    x.sort();
    let len = x.len();
    let x0 = x[1];
    let x1 = x0.borrow();

    println!("x.len = {len}, x[0]= {x0}, x1={x1}");
    let mut s = String::new();
    s.push_str("some");
    s.push_str("thing");
    s.push_str("went wrong");
    return s;
}

fn main() {
    let some = Some("something".to_string());
    let some_desc = some.string();

    print!("some_desc = {some_desc}");

    let mut vec = Vec::<u8>::new();
    unsafe {
        vec.append("some".parse::<String>().unwrap().as_mut_vec());
    }
    // let nbytes = some.write(vec).unwrap();

    let nbytes = "something went wrong"
        .parse::<String>()
        .unwrap()
        .write(vec)
        .unwrap();

    println!("Results: {nbytes} bytes written to some");

    {
        let x = foo4();
        println!("foo4 => '{x}'");
    }

    {
        let some = 30;
        _ = some;
    }

    {
        let _ = match "3.4e300".parse::<f64>() {
            Ok(num) => println!("num = {num}"),
            Err(err) => panic!("parse error: {err}"),
        };
    }
}
