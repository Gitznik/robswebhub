use actix_web::{get, HttpResponse, Responder};

#[get("/")]
async fn root() -> impl Responder {
    HttpResponse::Ok().body(include_str!("get.html"))
}
