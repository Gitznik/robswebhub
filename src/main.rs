use actix_web::{App, HttpServer};
use robswebhub::{configuration::get_configuration, routes::root::get::root};

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let config = get_configuration().unwrap();
    HttpServer::new(|| App::new().service(root))
        .bind((config.application.host, config.application.port))?
        .run()
        .await
}
