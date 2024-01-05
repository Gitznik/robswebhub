use actix_web::{get, head, HttpResponse, Responder};

use crate::html_base::compose_html;

#[get("/")]
async fn root() -> impl Responder {
    let main_div = include_str!("get.html");
    let html = compose_html(main_div);
    HttpResponse::Ok().body(html)
}

#[head("/")]
async fn root_head() -> impl Responder {
    HttpResponse::Ok()
}
