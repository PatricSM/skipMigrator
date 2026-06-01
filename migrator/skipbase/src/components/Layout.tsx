import { Outlet, useNavigate, Link, useLocation } from 'react-router-dom'
import { useAuth } from '@/hooks/use-auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarTrigger,
  SidebarInset,
} from '@/components/ui/sidebar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import {
  LayoutDashboard,
  FileText,
  Settings,
  Users,
  Search,
  Bell,
  LogOut,
  User,
} from 'lucide-react'

export default function Layout() {
  const { user, signOut } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  const handleSignOut = async () => {
    await signOut()
    navigate('/login')
  }

  const navItems = [
    { title: 'Painel de Controle', url: '/', icon: LayoutDashboard },
    { title: 'Relatórios', url: '/reports', icon: FileText },
    { title: 'Usuários', url: '/users', icon: Users },
    { title: 'Configurações', url: '/settings', icon: Settings },
  ]

  const getPageTitle = () => {
    const item = navItems.find((n) => n.url === location.pathname)
    return item ? item.title : 'Página'
  }

  return (
    <SidebarProvider>
      <Sidebar className="border-r shadow-sm">
        <SidebarHeader className="h-16 flex items-center justify-center border-b px-6">
          <div className="flex items-center gap-2 font-bold text-xl tracking-tight text-slate-900 dark:text-slate-100 w-full">
            <div className="bg-blue-600 p-1.5 rounded-lg flex items-center justify-center">
              <ActivityIcon className="text-white h-5 w-5" />
            </div>
            EticConecta
          </div>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu className="gap-2 mt-4 px-2">
                {navItems.map((item) => (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton
                      asChild
                      isActive={location.pathname === item.url}
                      className="data-[active=true]:bg-blue-50 data-[active=true]:text-blue-700 data-[active=true]:font-semibold hover:bg-slate-100"
                    >
                      <Link to={item.url}>
                        <item.icon className="h-5 w-5 mr-2" />
                        <span>{item.title}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter className="border-t p-4">
          <div className="flex items-center gap-3">
            <Avatar className="h-9 w-9 border">
              <AvatarFallback className="bg-slate-100 text-slate-600 font-medium">
                {user?.email?.substring(0, 2).toUpperCase() || 'US'}
              </AvatarFallback>
            </Avatar>
            <div className="flex flex-col overflow-hidden">
              <span className="text-sm font-medium truncate">{user?.email?.split('@')[0]}</span>
              <span className="text-xs text-slate-500 truncate">{user?.email}</span>
            </div>
          </div>
        </SidebarFooter>
      </Sidebar>

      <SidebarInset className="flex flex-col bg-slate-50/50 min-h-screen">
        <header className="h-16 flex items-center justify-between border-b bg-white px-4 md:px-6 sticky top-0 z-10">
          <div className="flex items-center gap-4">
            <SidebarTrigger />
            <div className="hidden md:flex">
              <Breadcrumb>
                <BreadcrumbList>
                  <BreadcrumbItem>
                    <BreadcrumbLink href="/">Plataforma</BreadcrumbLink>
                  </BreadcrumbItem>
                  <BreadcrumbSeparator />
                  <BreadcrumbItem>
                    <BreadcrumbPage>{getPageTitle()}</BreadcrumbPage>
                  </BreadcrumbItem>
                </BreadcrumbList>
              </Breadcrumb>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <div className="relative hidden md:block w-64">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-slate-400" />
              <Input
                type="search"
                placeholder="Pesquisar..."
                className="w-full bg-slate-50 pl-9 rounded-full border-slate-200 focus-visible:ring-blue-500"
              />
            </div>
            <Button variant="ghost" size="icon" className="text-slate-500 rounded-full">
              <Bell className="h-5 w-5" />
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="relative h-9 w-9 rounded-full">
                  <Avatar className="h-9 w-9 border border-slate-200">
                    <AvatarFallback className="bg-blue-100 text-blue-700 font-semibold">
                      {user?.email?.substring(0, 2).toUpperCase() || 'US'}
                    </AvatarFallback>
                  </Avatar>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56" align="end" forceMount>
                <DropdownMenuLabel className="font-normal">
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-medium leading-none">Minha Conta</p>
                    <p className="text-xs leading-none text-muted-foreground">{user?.email}</p>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem className="cursor-pointer">
                  <User className="mr-2 h-4 w-4" />
                  <span>Meu Perfil</span>
                </DropdownMenuItem>
                <DropdownMenuItem className="cursor-pointer">
                  <Settings className="mr-2 h-4 w-4" />
                  <span>Configurações</span>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  className="cursor-pointer text-red-600 focus:text-red-600"
                  onClick={handleSignOut}
                >
                  <LogOut className="mr-2 h-4 w-4" />
                  <span>Sair</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>

        <main className="flex-1 p-4 md:p-8 animate-fade-in-up">
          <Outlet />
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}

function ActivityIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
    </svg>
  )
}
