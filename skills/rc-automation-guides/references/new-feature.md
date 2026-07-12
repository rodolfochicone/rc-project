# Skill: /new-feature

Workflow guiado para criar uma nova feature end-to-end seguindo os padroes definidos nos `agent-docs/`.

## O que fazer ao invocar este skill

1. Leia os documentos de contexto necessarios:
   - `agent-docs/ARCHITECTURE.md` — camadas, DI, regras de importacao, checklist
   - `agent-docs/CODING-STANDARDS.md` — nomenclatura, TypeScript patterns, logging
   - `agent-docs/TESTING.md` — como escrever o teste co-localizado
   - `agent-docs/API.md` — se a feature expoe endpoint REST (rotas, paginacao, resposta)
   - `agent-docs/SECURITY.md` — validacao de inputs com Zod
   - `agent-docs/DATA-MODELING.md` — se a feature envolve novo schema de banco

2. Pergunte ao usuario:
   - Qual o nome/contexto da feature? (ex: `pricing`, `dealers`, `coverage`)
   - Qual o nome da entidade principal?
   - Quais operacoes o repositorio precisa expor?
   - A feature expoe endpoint HTTP, consome eventos SQS, ou ambos?
   - Precisa de nova tabela/migracao no banco?

3. Execute o checklist na ordem abaixo. Confirme cada passo antes de avancar para o proximo.

---

## Checklist de passos

### Passo 1 — Migracao de banco (se necessario)

Se a feature precisa de nova tabela ou alteracao de schema, crie a migracao primeiro.

Use `/db-migration` para o workflow completo, ou siga as regras minimas:
- Tabela: `snake_case`, plural, campos obrigatorios (`id UUID`, `created_at`, `updated_at`, `deleted_at`)
- Tipos corretos: `NUMERIC(15,2)` para monetarios, `TIMESTAMPTZ` para datas, `UUID` para IDs
- Indices para FKs e colunas de WHERE/ORDER BY
- Arquivo: `migrations/<timestamp>_<descricao-kebab-case>.sql`

---

### Passo 2 — Domain

Crie em `src/domain/<contexto>/`:
- Entidade TypeScript (`interface`, nunca `class` para dados simples)
- `<Nome>RepositoryInterface` — sempre `interface`, nunca `class` com `throw`
- `<Nome>RepositoryFactory` — seleciona implementacao via variavel de ambiente
- Domain Errors tipados se necessario (ex: `OrderNotFoundError`)

```typescript
/**
 * Entidade que representa um preco de produto.
 */
interface Price {
  id: string;
  productId: string;
  value: number;
  createdAt: string;
}

/**
 * Contrato do repositorio de precos.
 */
interface PriceRepositoryInterface {
  findAll(): Promise<Price[]>;
  findById(id: string): Promise<Price | null>;
  save(price: Price): Promise<void>;
}

/**
 * Factory que seleciona a implementacao do repositorio de precos.
 */
class PriceRepositoryFactory {
  static create({ infra, logger }: { infra: string; logger: ILogger }): PriceRepositoryInterface {
    if (infra === 'mock-file') return new PriceMockFileRepository(logger);
    if (infra === 'postgres') return new PostgresPriceRepository(logger);
    throw new Error(`Infraestrutura nao suportada: ${infra}`);
  }
}

export { Price, PriceRepositoryInterface, PriceRepositoryFactory };
```

Regras:
- Sem `any` — use `unknown` quando necessario
- Named exports — sem default exports
- Domain nao importa nada do projeto (apenas `@escaletech/logger`)

---

### Passo 3 — Infrastructure

Crie em `src/infrastructure/`:
- Implementacao concreta do repositorio (ex: `PostgresPriceRepository`, `PriceMockFileRepository`)
- Use `async/await` — nunca `new Promise` com callbacks
- Para mock-file: arquivo JSON em `src/infrastructure/mock-file/<contexto>/`
- Queries parametrizadas — nunca concatenar SQL com input

```typescript
/**
 * Repositorio de precos usando mock-file para testes.
 */
class PriceMockFileRepository implements PriceRepositoryInterface {
  constructor(private logger: ILogger) {}

  async findAll(): Promise<Price[]> {
    const raw = await fs.promises.readFile(DATA_FILE, 'utf8');
    return JSON.parse(raw);
  }
}
```

---

### Passo 4 — Usecase

Crie em `src/usecases/<contexto>/`:
- `<Nome>Service` com constructor injection de `ILogger` e repositorios
- Metodo publico `execute()` — um usecase = uma acao
- Logica de negocio aqui — nunca em controllers/handlers/consumers
- JSDoc descrevendo o que o usecase faz

```typescript
/**
 * Servico responsavel por listar precos de produtos.
 */
class ListPricesService {
  constructor(
    private logger: ILogger,
    private priceRepository: PriceRepositoryInterface
  ) {}

  /**
   * Lista todos os precos disponiveis.
   */
  async execute(): Promise<Price[]> {
    this.logger.info('Listando precos');
    return this.priceRepository.findAll();
  }
}
```

---

### Passo 5 — Teste co-localizado

Crie `<NomeService>.test.ts` na mesma pasta do service:
- **Abordagem preferida:** crie implementacao `InMemory*` que implementa a interface do dominio — testes deterministicos sem depender de infraestrutura
- **Alternativa:** use `infra: 'mock-file'` via Factory quando disponivel
- Nunca instancie o repositorio concreto (Postgres, HTTP) diretamente
- Use `beforeAll` para setup stateless

```typescript
// Abordagem preferida: InMemory implementation
class InMemoryPriceRepository implements PriceRepositoryInterface {
  private prices: Price[] = [
    { id: '1', productId: 'p1', value: 99.90, createdAt: new Date().toISOString() }
  ];
  async findAll(): Promise<Price[]> { return this.prices; }
  async findById(id: string): Promise<Price | null> {
    return this.prices.find(p => p.id === id) ?? null;
  }
  async save(price: Price): Promise<void> { this.prices.push(price); }
}

describe('ListPricesService', () => {
  let service: ListPricesService;

  beforeAll(() => {
    service = new ListPricesService(logger, new InMemoryPriceRepository());
  });

  it('deve retornar array de precos', async () => {
    const prices = await service.execute();
    expect(prices).toBeInstanceOf(Array);
    expect(prices.length).toBeGreaterThan(0);
  });
});
```

---

### Passo 6 — Dependencies (server)

Adicione lazy singleton em `src/interfaces/server/dependencies.ts`:

```typescript
private static _listPricesService: ListPricesService;

static getListPricesService(): ListPricesService {
  if (!Dependencies._listPricesService) {
    Dependencies._listPricesService = new ListPricesService(
      Dependencies.getLogger(),
      Dependencies.getPriceRepository()
    );
  }
  return Dependencies._listPricesService;
}
```

---

### Passo 7 — Dependencies (lambda)

Adicione o MESMO lazy singleton em `src/interfaces/lambda/dependencies.ts`.
**ATENCAO:** Este passo e obrigatorio. Os dois arquivos devem estar sincronizados.

---

### Passo 8 — Controller (server) e/ou Handler (lambda)

**Controller** em `src/interfaces/server/controllers/`:
- Arrow methods para binding correto do `this`
- Validacao de input com Zod na fronteira
- Delegue para o usecase — sem logica de negocio aqui
- JSDoc no controller

```typescript
import { z } from 'zod';

const ListPricesQuerySchema = z.object({
  page: z.coerce.number().int().positive().optional().default(1),
  limit: z.coerce.number().int().positive().max(100).optional().default(20),
});

/**
 * Controller responsavel pelos endpoints de precos.
 */
export class PriceController {
  constructor(private listPricesService: ListPricesService) {}

  listPrices = async (req: XclRequest, res: Response): Promise<void> => {
    const query = ListPricesQuerySchema.safeParse(req.query);
    if (!query.success) {
      const details = query.error.issues.map(i => ({ field: i.path.join('.'), message: i.message }));
      res.status(400).json({ error: { code: 'VALIDATION_ERROR', message: 'Request validation failed.', details } });
      return;
    }
    const prices = await this.listPricesService.execute();
    res.status(200).json(prices);
  };
}
```

**Handler** em `src/interfaces/lambda/<contexto>/`:
- Importe `Context` de `aws-lambda` — nunca use `context` sem tipo
- Defina correlation ID: `logger.setCorrectionId(context.awsRequestId)`
- Validacao de input com Zod

```typescript
import { Context, APIGatewayProxyEvent } from 'aws-lambda';

/**
 * Handler Lambda para listar precos.
 */
export const handler = async (event: APIGatewayProxyEvent, context: Context) => {
  const logger = Dependencies.getLogger();
  logger.setCorrectionId(context.awsRequestId);
  // validacao com Zod + delegacao ao usecase
};
```

---

### Passo 9 — Consumer SQS (se a feature consome eventos)

Crie em `src/interfaces/sqs/consumers/`:
- Consumer delega ao usecase — sem logica de negocio
- Deve ser idempotente
- Validar body com Zod
- Definir correlation ID

```typescript
/**
 * Consumer para processar eventos de price.updated.
 */
export class PriceUpdatedConsumer {
  constructor(private readonly updatePrice: UpdatePriceService) {}

  async handle(message: SQSMessage): Promise<void> {
    const input = PriceUpdatedSchema.parse(JSON.parse(message.Body));
    await this.updatePrice.execute(input);
  }
}
```

---

### Passo 10 — serverless.yaml

Registre a nova funcao no arquivo `serverless.yaml`:

```yaml
# Handler para listar precos de produtos
functions:
  listPrices:
    handler: src/interfaces/lambda/pricing/listPrices.handler
    events:
      - http:
          path: /energy/pricing
          method: GET
          integration: lambda
          response: ${file(config/response.yaml)}
          request: ${file(config/request.yaml)}
```

Registre tambem em `serverless.dev.yaml` para testes locais com `serverless-offline`.

---

### Passo 11 — OpenAPI

Atualize `/docs/openapi.yaml` com:
- Endpoint, parametros e responses documentados
- Schemas de request body e response com tipos definidos
- Codigos de erro possiveis por endpoint
- Headers de contexto (`partner`, `product`) documentados

---

### Passo 12 — Publicacao de evento (se a feature publica eventos)

Se o usecase precisa notificar outros servicos:

```typescript
await publisher.publishEvent([{
  object: 'price',
  action: 'created',
  state: 'active',
  body: { id: price.id, value: price.value },
  coi: price.id,
  partner: context.partner,
  product: context.product,
  operation: context.operation,
  source: 'pricing-service',
}]);
```

Regras: publicar **apos** persistencia, sempre propagar contexto, body enxuto, sem dados sensiveis.

---

## Verificacao final

Apos completar todos os passos aplicaveis, execute:

```bash
pnpm test    # deve passar com 0 falhas
pnpm build   # deve compilar sem erros
```

Se qualquer um falhar, corrija antes de considerar a feature completa.

### Checklist de revisao rapida

- [ ] Regras de importacao entre camadas respeitadas?
- [ ] `dependencies.ts` sincronizado em server/ e lambda/?
- [ ] Inputs validados com Zod na fronteira?
- [ ] Testes co-localizados com InMemory* (preferido) ou mock-file via Factory?
- [ ] JSDoc nas classes e funcoes relevantes?
- [ ] Sem `console.*`, sem `any`, sem default exports?
- [ ] OpenAPI atualizado (se endpoint HTTP)?
- [ ] Comentarios no handler do serverless.yaml?
