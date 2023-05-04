use actix_web::{App, HttpServer};
use robswebhub::{
    configuration::get_configuration,
    routes::{about::get::about, root::get::root, scores::get::add_scores},
};

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let config = get_configuration().unwrap();
    HttpServer::new(|| {
        App::new()
            .service(root)
            .service(about)
            .service(add_scores)
    })
    .bind((config.application.host, config.application.port))?
    .run()
    .await
}

