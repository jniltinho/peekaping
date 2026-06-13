## ADDED Requirements

### Requirement: Design tokens Neo-Brutalism em CSS
O sistema SHALL definir tokens CSS Neo-Brutalism em `apps/web/src/index.css` via `@theme inline`, incluindo:
- `--shadow-neo`: sombra hard-offset de 4px (`4px 4px 0px 0px oklch(0 0 0)`)
- `--shadow-neo-sm`: sombra hard-offset de 2px (`2px 2px 0px 0px oklch(0 0 0)`)
- `--border` definido como preto puro (`oklch(0 0 0)`) no light mode
- No dark mode, sombras e bordas SHALL usar branco puro (`oklch(1 0 0)`)
- `--radius` SHALL ser `0px` (sem arredondamento em nenhum elemento)

#### Scenario: Tokens disponíveis como classes Tailwind
- **WHEN** o CSS é compilado pelo Tailwind v4
- **THEN** as classes `shadow-neo` e `shadow-neo-sm` estão disponíveis para uso em qualquer componente

#### Scenario: Dark mode inverte cores das sombras
- **WHEN** o tema dark está ativo (classe `.dark` no elemento raiz)
- **THEN** sombras e bordas usam branco puro para manter alto contraste

#### Scenario: Nenhum elemento tem border-radius
- **WHEN** qualquer componente UI é renderizado
- **THEN** nenhum elemento exibe cantos arredondados (`border-radius: 0` em todo o sistema)

### Requirement: Componentes UI com bordas e sombras Neo-Brutalism
Os componentes em `apps/web/src/components/ui/` SHALL aplicar por padrão:
- Borda preta de 2px (`border-2 border-black`) em: `Button` (variantes default, outline, secondary), `Card`, `Input`, `Textarea`, `Select`, `Badge`, `Alert`, `Dialog`, `Sheet`
- Sombra `shadow-neo` em: `Card`, `Button` (default), `Dialog`, `Sheet`
- Sombra `shadow-neo-sm` em: `Input`, `Textarea`, `Select`, `Badge`
- Todos os `rounded-*` SHALL ser removidos ou sobrescritos por `rounded-none`

#### Scenario: Button padrão exibe borda e sombra
- **WHEN** um `<Button>` com variant="default" é renderizado
- **THEN** exibe borda preta de 2px e sombra hard-offset de 4px

#### Scenario: Card exibe bordas grossas e sombra deslocada
- **WHEN** um `<Card>` é renderizado
- **THEN** exibe borda preta de 2px e sombra `4px 4px 0px 0px black`

#### Scenario: Input exibe borda preta sem arredondamento
- **WHEN** um `<Input>` ou `<Textarea>` é renderizado
- **THEN** exibe borda preta de 2px, sem border-radius e sem sombra de foco padrão

#### Scenario: Badge exibe borda e sombra pequena
- **WHEN** um `<Badge>` é renderizado
- **THEN** exibe borda preta de 2px e sombra `shadow-neo-sm`

### Requirement: Componentes globais com visual Neo-Brutalism
Os componentes globais em `apps/web/src/components/` SHALL aplicar estilo Neo-Brutalism:
- `app-sidebar.tsx`: borda direita preta de 2px, sem sombra lateral suave
- `app-header.tsx`: borda inferior preta de 2px
- `nav-main.tsx`: itens ativos com fundo de destaque e borda esquerda preta de 4px
- `nav-user.tsx`: borda preta de 2px no dropdown

#### Scenario: Sidebar com borda direita visível
- **WHEN** a sidebar é renderizada
- **THEN** exibe borda direita preta sólida de 2px sem box-shadow lateral suave

#### Scenario: Header com borda inferior
- **WHEN** o header da aplicação é renderizado
- **THEN** exibe borda inferior preta sólida de 2px

#### Scenario: Item ativo no nav com borda esquerda de destaque
- **WHEN** um item de navegação está na rota ativa
- **THEN** exibe borda esquerda preta de 4px como indicador visual

### Requirement: Hover com efeito de deslocamento de sombra
Elementos interativos (Button, Card clicável) SHALL implementar efeito de hover Neo-Brutalism: ao hover, a sombra é reduzida e o elemento se desloca levemente (translate), simulando profundidade.

#### Scenario: Button hover desloca visualmente
- **WHEN** o usuário passa o cursor sobre um Button
- **THEN** o elemento aplica `translate-x-[2px] translate-y-[2px]` e reduz a sombra para `shadow-neo-sm`

#### Scenario: Transição suave no hover
- **WHEN** o cursor entra e sai de um elemento interativo
- **THEN** a transição ocorre com `transition-all duration-100` (rápida, característica Neo-Brutalism)
