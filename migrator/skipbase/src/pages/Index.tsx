import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Users,
  Activity,
  CheckCircle2,
  Clock,
  ArrowUpRight,
  ArrowDownRight,
  UserPlus,
  Download,
} from 'lucide-react'
import {
  Area,
  AreaChart,
  CartesianGrid,
  XAxis,
  YAxis,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
} from 'recharts'
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

// Mock Data for Charts
const chartData = [
  { date: 'Segunda', accesses: 120, operations: 80 },
  { date: 'Terça', accesses: 210, operations: 130 },
  { date: 'Quarta', accesses: 180, operations: 110 },
  { date: 'Quinta', accesses: 290, operations: 190 },
  { date: 'Sexta', accesses: 240, operations: 150 },
  { date: 'Sábado', accesses: 90, operations: 40 },
  { date: 'Domingo', accesses: 110, operations: 60 },
]

const chartConfig = {
  accesses: {
    label: 'Acessos',
    color: 'hsl(var(--chart-1))',
  },
  operations: {
    label: 'Operações',
    color: 'hsl(var(--chart-2))',
  },
}

// Mock Data for Table
const recentActivity = [
  {
    id: '1',
    action: 'Novo Login',
    user: 'patric.martins@adapta.org',
    time: 'Hoje, 10:23',
    status: 'Sucesso',
  },
  {
    id: '2',
    action: 'Exportação de Relatório',
    user: 'ana.silva@empresa.com',
    time: 'Hoje, 09:15',
    status: 'Sucesso',
  },
  {
    id: '3',
    action: 'Falha de Autenticação',
    user: 'desconhecido',
    time: 'Ontem, 14:12',
    status: 'Erro',
  },
  {
    id: '4',
    action: 'Atualização de Perfil',
    user: 'carlos.mendes@tec.br',
    time: 'Ontem, 11:30',
    status: 'Aviso',
  },
  {
    id: '5',
    action: 'Criação de Usuário',
    user: 'patric.martins@adapta.org',
    time: '24/05/2026, 16:45',
    status: 'Sucesso',
  },
]

export default function Index() {
  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h2 className="text-3xl font-bold tracking-tight text-slate-900">Visão Geral</h2>
          <p className="text-slate-500">Bem-vindo ao seu painel de controle administrativo.</p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" className="bg-white">
            <Download className="mr-2 h-4 w-4" />
            Exportar Dados
          </Button>
          <Button className="bg-blue-600 hover:bg-blue-700 text-white">
            <UserPlus className="mr-2 h-4 w-4" />
            Convidar Usuário
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card className="hover:shadow-md transition-shadow">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-600">Total de Usuários</CardTitle>
            <Users className="h-4 w-4 text-blue-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-slate-900">1,248</div>
            <p className="text-xs text-emerald-600 flex items-center mt-1">
              <ArrowUpRight className="h-3 w-3 mr-1" />
              +12% no último mês
            </p>
          </CardContent>
        </Card>
        <Card className="hover:shadow-md transition-shadow">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-600">Sessões Ativas</CardTitle>
            <Activity className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-slate-900">142</div>
            <p className="text-xs text-emerald-600 flex items-center mt-1">
              <ArrowUpRight className="h-3 w-3 mr-1" />
              +4% em relação a ontem
            </p>
          </CardContent>
        </Card>
        <Card className="hover:shadow-md transition-shadow">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-600">Saúde do Sistema</CardTitle>
            <CheckCircle2 className="h-4 w-4 text-emerald-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-slate-900">99.9%</div>
            <p className="text-xs text-slate-500 flex items-center mt-1">
              <span className="h-2 w-2 rounded-full bg-emerald-500 mr-2"></span>
              Operando normalmente
            </p>
          </CardContent>
        </Card>
        <Card className="hover:shadow-md transition-shadow">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-600">Tarefas Pendentes</CardTitle>
            <Clock className="h-4 w-4 text-rose-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-slate-900">12</div>
            <p className="text-xs text-rose-600 flex items-center mt-1">
              <ArrowDownRight className="h-3 w-3 mr-1" />
              -2% no tempo de resolução
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-7">
        <Card className="col-span-1 md:col-span-4 lg:col-span-5 hover:shadow-md transition-shadow">
          <CardHeader>
            <CardTitle>Atividade Recente</CardTitle>
            <CardDescription>
              Volume de acessos e operações processadas nos últimos 7 dias.
            </CardDescription>
          </CardHeader>
          <CardContent className="pl-0">
            <ChartContainer config={chartConfig} className="h-[300px] w-full mt-4">
              <AreaChart
                accessibilityLayer
                data={chartData}
                margin={{
                  left: 12,
                  right: 12,
                  top: 10,
                  bottom: 0,
                }}
              >
                <defs>
                  <linearGradient id="fillAccesses" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="var(--color-accesses)" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="var(--color-accesses)" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="fillOperations" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="var(--color-operations)" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="var(--color-operations)" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e2e8f0" />
                <XAxis
                  dataKey="date"
                  tickLine={false}
                  axisLine={false}
                  tickMargin={8}
                  tick={{ fill: '#64748b', fontSize: 12 }}
                />
                <YAxis tickLine={false} axisLine={false} tick={{ fill: '#64748b', fontSize: 12 }} />
                <ChartTooltip cursor={false} content={<ChartTooltipContent />} />
                <Area
                  type="monotone"
                  dataKey="operations"
                  stroke="var(--color-operations)"
                  fill="url(#fillOperations)"
                  strokeWidth={2}
                  stackId="2"
                />
                <Area
                  type="monotone"
                  dataKey="accesses"
                  stroke="var(--color-accesses)"
                  fill="url(#fillAccesses)"
                  strokeWidth={2}
                  stackId="1"
                />
              </AreaChart>
            </ChartContainer>
          </CardContent>
        </Card>

        <Card className="col-span-1 md:col-span-3 lg:col-span-2 hover:shadow-md transition-shadow flex flex-col">
          <CardHeader>
            <CardTitle>Ações Rápidas</CardTitle>
            <CardDescription>Atalhos para tarefas comuns.</CardDescription>
          </CardHeader>
          <CardContent className="flex-1 flex flex-col gap-3">
            <Button variant="outline" className="w-full justify-start font-normal h-12">
              <UserPlus className="mr-3 h-4 w-4 text-blue-500" />
              Convidar Novo Usuário
            </Button>
            <Button variant="outline" className="w-full justify-start font-normal h-12">
              <FileText className="mr-3 h-4 w-4 text-emerald-500" />
              Gerar Relatório Mensal
            </Button>
            <Button variant="outline" className="w-full justify-start font-normal h-12">
              <Settings className="mr-3 h-4 w-4 text-slate-500" />
              Configurações de Segurança
            </Button>

            <div className="mt-auto pt-6">
              <div className="bg-blue-50 rounded-lg p-4 border border-blue-100">
                <h4 className="text-sm font-semibold text-blue-900 mb-1">Atualização Disponível</h4>
                <p className="text-xs text-blue-700 mb-3">
                  Uma nova versão do sistema está pronta para instalação.
                </p>
                <Button size="sm" className="w-full bg-blue-600 hover:bg-blue-700 text-white">
                  Instalar Agora
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card className="hover:shadow-md transition-shadow">
        <CardHeader>
          <CardTitle>Logs de Sistema</CardTitle>
          <CardDescription>
            Histórico das últimas operações realizadas na plataforma.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead>Ação</TableHead>
                <TableHead>Usuário</TableHead>
                <TableHead>Data/Hora</TableHead>
                <TableHead className="text-right">Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {recentActivity.map((activity) => (
                <TableRow key={activity.id}>
                  <TableCell className="font-medium text-slate-900">{activity.action}</TableCell>
                  <TableCell className="text-slate-600">{activity.user}</TableCell>
                  <TableCell className="text-slate-500">{activity.time}</TableCell>
                  <TableCell className="text-right">
                    <Badge
                      variant="outline"
                      className={
                        activity.status === 'Sucesso'
                          ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                          : activity.status === 'Erro'
                            ? 'bg-rose-50 text-rose-700 border-rose-200'
                            : 'bg-amber-50 text-amber-700 border-amber-200'
                      }
                    >
                      {activity.status}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
