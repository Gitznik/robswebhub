pub fn compose_html(main_div: &str) -> String {
    let header = include_str!("header.html");
    let navbar = include_str!("navbar.html");
    let footer = include_str!("footer.html");
    let html = format!(r#"
<!doctype html>
<html lang="en" data-theme="dark">
{}
  <body>
    {}
    {}
  </body>
</html>
{}
    "#, header, navbar, main_div, footer);
    html
} 

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn composing_works() {
        let html = compose_html("");
        dbg!(&html);
        assert!(html.contains("favicon"));
    }

}
