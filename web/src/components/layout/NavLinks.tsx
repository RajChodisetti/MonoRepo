import Link from "next/link";
import NavDropdown, { type NavDropdownItem } from "@/components/layout/NavDropdown";
import ProductMegaMenu from "@/components/layout/ProductMegaMenu";

export type NavLink =
  | { type: "link"; label: string; href: string }
  | { type: "dropdown"; label: string; items: NavDropdownItem[] }
  | { type: "mega"; label: "Product" };

type NavLinksProps = {
  links: NavLink[];
  className?: string;
};

export default function NavLinks({ links, className = "" }: NavLinksProps) {
  return (
    <ul className={`items-center gap-12 xl:gap-14 ${className}`}>
      {links.map((item) => (
        <li key={item.label}>
          {item.type === "mega" ? (
            <ProductMegaMenu />
          ) : item.type === "dropdown" ? (
            <NavDropdown label={item.label} items={item.items} />
          ) : (
            <Link
              href={item.href}
              className="text-[15px] font-semibold text-[#1a1a1a] transition-colors hover:text-black"
            >
              {item.label}
            </Link>
          )}
        </li>
      ))}
    </ul>
  );
}
