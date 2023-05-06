use actix_web::{App, HttpServer};
use robswebhub::{
    configuration::get_configuration,
    routes::{about::get::about, root::get::root, scores::{get::add_scores, post::save_scores}},
};
use actix_files::Files;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let config = get_configuration().unwrap();
    HttpServer::new(|| {
        App::new()
            .service(root)
            .service(about)
            .service(add_scores)
            .service(save_scores)
            .service(Files::new("/images", "./images"))
    })
    .bind((config.application.host, config.application.port))?
    .run()
    .await
}

