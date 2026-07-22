import { foodieContent } from "./lib/foodieContent";
import FoodieNav from "./components/FoodieNav";
import FoodieHero from "./components/FoodieHero";
import FoodieMenu from "./components/FoodieMenu";
import FoodieReviews from "./components/FoodieReviews";
import FoodieCta from "./components/FoodieCta";
import FoodieContact from "./components/FoodieContact";
import FoodieFooter from "./components/FoodieFooter";
import "./theme.css";

export default function FoodieTemplate() {
  const content = foodieContent;

  return (
    <div
      className="foodie-root"
      style={{ fontFamily: "var(--font-foodie-body), Poppins, system-ui, sans-serif" }}
    >
      <FoodieNav brand={content.brand} links={content.nav} />
      <main>
        <section id="home">
          <FoodieHero hero={content.hero} dish={content.dish} />
        </section>
        <FoodieMenu menu={content.menu} />
        <FoodieReviews reviews={content.reviews} />
        <FoodieCta cta={content.cta} />
        <FoodieContact contact={content.contact} brandName={content.brand.name} />
      </main>
      <FoodieFooter brand={content.brand} footer={content.footer} contact={content.contact} />
    </div>
  );
}
