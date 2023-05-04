use actix_web::{get, HttpResponse, Responder};

#[get("/about")]
async fn about() -> impl Responder {
    HttpResponse::Ok().body(include_str!("get.html"))
}
