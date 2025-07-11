import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '../ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu';
import useAuth from '../../hooks/useAuth';

function Header() {
  const { isLoggedIn, user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
  };

  return (
    <header className="bg-primary text-primary-foreground p-4 shadow-md flex justify-center items-center sticky top-0 z-50">
      <div className="header-content w-full px-4 flex justify-between items-center">
        <h1 className="text-2xl font-bold">
          <Link to="/dashboard" className="text-primary-foreground no-underline hover:text-primary-foreground/80">
            Tiketmu
          </Link>
        </h1>
        <nav>
          <ul className="flex items-center space-x-6">
            <li><Link to="/dashboard" className="text-primary-foreground hover:text-primary-foreground/80">Concerts</Link></li>
            {isLoggedIn ? (
              <>
                <li><Link to="/my-bookings" className="text-primary-foreground hover:text-primary-foreground/80">My Bookings</Link></li>
                <li>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" className="text-primary-foreground hover:bg-primary/80 px-3 py-2 rounded-md">
                        Hi, {user?.username || 'User'}
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent className="w-48 bg-card text-card-foreground shadow-lg rounded-md mt-1 border border-border">
                      <DropdownMenuLabel className="px-3 py-2 text-sm font-semibold text-foreground">My Account</DropdownMenuLabel>
                      <DropdownMenuSeparator className="border-t border-border" />
                      <DropdownMenuItem className="cursor-pointer px-3 py-2 text-foreground hover:bg-accent" onClick={() => navigate('/profile')}>
                        Edit Profile
                      </DropdownMenuItem>
                      <DropdownMenuItem className="cursor-pointer px-3 py-2 text-foreground hover:bg-accent" onClick={handleLogout}>
                        Logout
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </li>
              </>
            ) : (
              <li><Link to="/auth" className="text-primary-foreground hover:text-primary-foreground/80">Login / Register</Link></li>
            )}
          </ul>
        </nav>
      </div>
    </header>
  );
}

export default Header;