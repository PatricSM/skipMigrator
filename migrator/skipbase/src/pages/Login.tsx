import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import { useAuth } from '@/hooks/use-auth'
import { useToast } from '@/hooks/use-toast'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Checkbox } from '@/components/ui/checkbox'
import { Loader2 } from 'lucide-react'

const formSchema = z.object({
  email: z.string().email({ message: 'E-mail inválido.' }),
  password: z.string().min(6, { message: 'A senha deve ter pelo menos 6 caracteres.' }),
  remember: z.boolean().default(false).optional(),
})

export default function Login() {
  const [isLoading, setIsLoading] = useState(false)
  const { signIn } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const { toast } = useToast()

  const from = location.state?.from?.pathname || '/'

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      email: '',
      password: '',
      remember: false,
    },
  })

  async function onSubmit(values: z.infer<typeof formSchema>) {
    setIsLoading(true)

    const { error } = await signIn(values.email, values.password)

    setIsLoading(false)

    if (error) {
      toast({
        variant: 'destructive',
        title: 'Falha na autenticação',
        description: 'Verifique seu e-mail e senha e tente novamente.',
      })
      return
    }

    toast({
      title: 'Bem-vindo(a) de volta!',
      description: 'Login realizado com sucesso.',
    })

    navigate(from, { replace: true })
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-50 px-4 relative overflow-hidden">
      {/* Decorative background blobs */}
      <div className="absolute top-0 -left-4 w-72 h-72 bg-blue-300 rounded-full mix-blend-multiply filter blur-2xl opacity-30 animate-float"></div>
      <div
        className="absolute top-0 -right-4 w-72 h-72 bg-emerald-300 rounded-full mix-blend-multiply filter blur-2xl opacity-30 animate-float"
        style={{ animationDelay: '2s' }}
      ></div>
      <div
        className="absolute -bottom-8 left-20 w-72 h-72 bg-amber-300 rounded-full mix-blend-multiply filter blur-2xl opacity-30 animate-float"
        style={{ animationDelay: '4s' }}
      ></div>

      <div className="w-full max-w-md z-10 animate-fade-in-up">
        <div className="text-center mb-8">
          <div className="mx-auto bg-blue-600 h-12 w-12 rounded-xl flex items-center justify-center mb-4 shadow-lg">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="white"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
            </svg>
          </div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-900">EticConecta</h1>
          <p className="text-slate-500 mt-2">Plataforma Administrativa Profissional</p>
        </div>

        <Card className="border-none shadow-xl shadow-slate-200/50">
          <CardHeader>
            <CardTitle className="text-xl">Entrar</CardTitle>
            <CardDescription>Insira suas credenciais para acessar sua conta.</CardDescription>
          </CardHeader>
          <CardContent>
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                <FormField
                  control={form.control}
                  name="email"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>E-mail</FormLabel>
                      <FormControl>
                        <Input placeholder="seu@email.com" {...field} className="bg-slate-50/50" />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="password"
                  render={({ field }) => (
                    <FormItem>
                      <div className="flex items-center justify-between">
                        <FormLabel>Senha</FormLabel>
                        <a
                          href="#"
                          className="text-sm font-medium text-blue-600 hover:text-blue-500"
                        >
                          Esqueceu a senha?
                        </a>
                      </div>
                      <FormControl>
                        <Input
                          type="password"
                          placeholder="••••••••"
                          {...field}
                          className="bg-slate-50/50"
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="remember"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-start space-x-3 space-y-0 rounded-md py-2">
                      <FormControl>
                        <Checkbox checked={field.value} onCheckedChange={field.onChange} />
                      </FormControl>
                      <div className="space-y-1 leading-none">
                        <FormLabel className="text-sm font-normal text-slate-600">
                          Lembrar de mim
                        </FormLabel>
                      </div>
                    </FormItem>
                  )}
                />
                <Button
                  type="submit"
                  className="w-full bg-blue-600 hover:bg-blue-700 text-white shadow-md shadow-blue-500/20"
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Autenticando...
                    </>
                  ) : (
                    'Entrar na Plataforma'
                  )}
                </Button>
              </form>
            </Form>
          </CardContent>
          <CardFooter className="flex justify-center border-t p-4 bg-slate-50/50 rounded-b-xl">
            <p className="text-sm text-slate-500">
              Não tem uma conta?{' '}
              <a href="#" className="font-medium text-blue-600 hover:text-blue-500">
                Solicite acesso
              </a>
            </p>
          </CardFooter>
        </Card>
      </div>
    </div>
  )
}
