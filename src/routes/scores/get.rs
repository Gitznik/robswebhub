use actix_web::{get, HttpResponse, Responder};

use crate::html_base::compose_html;

#[get("/scores")]
async fn add_scores() -> impl Responder {
    let main_div = include_str!("get.html");
    let html = compose_html(main_div);
    HttpResponse::Ok().body(html)
}
