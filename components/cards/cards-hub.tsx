"use client";

import { browserApiFetch } from "@/lib/api/browser-api";
import type { CreditCard, FinancialInstitution } from "@/lib/api/types";
import {
  Alert,
  Button,
  Card,
  CardContent,
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerTitle,
  Input,
  Spinner,
} from "@lucaslfs2004/luke-ui";
import {
  Archive,
  ArrowRight,
  CalendarDays,
  Check,
  CreditCard as CreditCardIcon,
  Landmark,
  Pencil,
  ReceiptText,
  Save,
  Wifi,
  X,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState, type FormEvent } from "react";

type Props = { institutions: FinancialInstitution[]; cards: CreditCard[] };
type Feedback = { type: "success" | "error"; message: string } | null;

function institutionTheme(name: string) {
  const normalized = name
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase();
  if (normalized.includes("nubank") || normalized === "nu") return "nubank";
  if (normalized.includes("itau")) return "itau";
  if (normalized.includes("mercado pago") || normalized.includes("mercadopago"))
    return "mercado-pago";
  return "default";
}

export function CardsHub({ institutions, cards }: Props) {
  const router = useRouter();
  const [drawer, setDrawer] = useState<"institution" | "card" | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [institutionPending, setInstitutionPending] = useState(false);
  const [cardPending, setCardPending] = useState(false);
  const [pendingAction, setPendingAction] = useState<string | null>(null);
  const [institutionFeedback, setInstitutionFeedback] =
    useState<Feedback>(null);
  const [cardFeedback, setCardFeedback] = useState<Feedback>(null);
  const [actionError, setActionError] = useState("");
  const [editing, setEditing] = useState<string | null>(null);
  const activeInstitutions = institutions.filter(
    (institution) => institution.status === "active",
  );

  useEffect(() => {
    const openFromHash = () => {
      if (window.location.hash === "#new-card") {
        setDrawer("card");
        setDrawerOpen(true);
      }
      if (window.location.hash === "#new-institution") {
        setDrawer("institution");
        setDrawerOpen(true);
      }
    };
    openFromHash();
    window.addEventListener("hashchange", openFromHash);
    return () => window.removeEventListener("hashchange", openFromHash);
  }, []);

  function closeDrawer() {
    if (institutionPending || cardPending) return;
    setDrawerOpen(false);
    if (
      window.location.hash === "#new-card" ||
      window.location.hash === "#new-institution"
    ) {
      window.history.replaceState(
        null,
        "",
        window.location.pathname + window.location.search,
      );
    }
  }

  async function submitInstitution(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const name = String(new FormData(form).get("name") ?? "").trim();
    setInstitutionPending(true);
    setInstitutionFeedback(null);
    try {
      await browserApiFetch("/financial-institutions", {
        method: "POST",
        body: JSON.stringify({ name }),
      });
      form.reset();
      setInstitutionFeedback({
        type: "success",
        message: `${name} cadastrada com sucesso.`,
      });
      setDrawerOpen(false);
      router.refresh();
    } catch (error) {
      setInstitutionFeedback({
        type: "error",
        message:
          error instanceof Error
            ? error.message
            : "Não foi possível cadastrar a instituição.",
      });
    } finally {
      setInstitutionPending(false);
    }
  }

  async function submitCard(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const values = new FormData(form);
    const name = String(values.get("name") ?? "").trim();
    const payload = {
      institution_id: String(values.get("institution_id")),
      name,
      configuration: {
        effective_from: String(values.get("effective_from")),
        end_month: String(values.get("end_month") || "") || null,
        nominal_due_day: Number(values.get("nominal_due_day")),
        payment_month_offset: Number(values.get("payment_month_offset")),
        context: String(values.get("context") || "").trim() || null,
      },
    };
    setCardPending(true);
    setCardFeedback(null);
    try {
      await browserApiFetch("/credit-cards", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      form.reset();
      setCardFeedback({
        type: "success",
        message: `Cartão ${name} cadastrado com sucesso.`,
      });
      setDrawerOpen(false);
      router.refresh();
    } catch (error) {
      setCardFeedback({
        type: "error",
        message:
          error instanceof Error
            ? error.message
            : "Não foi possível cadastrar o cartão.",
      });
    } finally {
      setCardPending(false);
    }
  }

  async function action(
    cardId: string,
    path: string,
    method: string,
    body?: unknown,
  ) {
    setPendingAction(cardId);
    setActionError("");
    try {
      await browserApiFetch(path, {
        method,
        body: body ? JSON.stringify(body) : undefined,
      });
      setEditing(null);
      router.refresh();
    } catch (error) {
      setActionError(
        error instanceof Error
          ? error.message
          : "Não foi possível concluir a operação.",
      );
    } finally {
      setPendingAction(null);
    }
  }

  return (
    <div className="cards-hub">
      <div
        className="grid grid-cols-2 gap-2 md:gap-3"
        aria-label="Novos cadastros"
      >
        <Button
          id="new-card"
          variant="ghost"
          className="luke-surface-glass w-full min-h-12 gap-1.5 border border-glass-border px-2 text-[11px] text-primary hover:bg-glass-background sm:text-xs md:min-h-14 md:gap-2.5 md:px-4 md:text-sm"
          type="button"
          onClick={() => {
            setCardFeedback(null);
            setDrawer("card");
            setDrawerOpen(true);
          }}
        >
          <CreditCardIcon size={19} className="shrink-0" />
          Novo cartão
        </Button>
        <Button
          id="new-institution"
          variant="ghost"
          className="luke-surface-glass w-full min-h-12 gap-1.5 border border-glass-border px-2 text-[11px] text-primary hover:bg-glass-background sm:text-xs md:min-h-14 md:gap-2.5 md:px-4 md:text-sm"
          type="button"
          onClick={() => {
            setInstitutionFeedback(null);
            setDrawer("institution");
            setDrawerOpen(true);
          }}
        >
          <Landmark size={19} className="shrink-0" />
          Nova instituição
        </Button>
      </div>
      <div className="cards-create-feedback" aria-live="polite">
        {cardFeedback?.type === "success" && (
          <FormFeedback
            pending={false}
            pendingLabel=""
            feedback={cardFeedback}
          />
        )}
        {institutionFeedback?.type === "success" && (
          <FormFeedback
            pending={false}
            pendingLabel=""
            feedback={institutionFeedback}
          />
        )}
      </div>

      <Drawer
        open={drawerOpen}
        onOpenChange={(open) => {
          if (!open) closeDrawer();
        }}
      >
        <DrawerContent
          side="right"
          className="cards-drawer"
          hideClose
          onEscapeKeyDown={(event) => {
            if (institutionPending || cardPending) event.preventDefault();
          }}
          onInteractOutside={(event) => {
            if (institutionPending || cardPending) event.preventDefault();
          }}
        >
          {drawer && (
            <>
              <div className="cards-drawer-header">
                <div>
                  <span className="section-kicker">
                    {drawer === "institution" ? "INSTITUIÇÕES" : "CARTÕES"}
                  </span>
                  <DrawerTitle id="cards-drawer-title">
                    {drawer === "institution"
                      ? "Nova instituição"
                      : "Novo cartão"}
                  </DrawerTitle>
                  <DrawerDescription className="sr-only">
                    {drawer === "institution"
                      ? "Formulário para cadastrar uma instituição financeira."
                      : "Formulário para cadastrar um cartão de crédito."}
                  </DrawerDescription>
                </div>
                <button
                  type="button"
                  className="cards-drawer-close"
                  aria-label="Fechar"
                  onClick={closeDrawer}
                  disabled={institutionPending || cardPending}
                >
                  <X size={20} />
                </button>
              </div>
              <div className="cards-drawer-body">
                {drawer === "institution" ? (
                  <>
                    <p className="section-description">
                      Cadastre o banco ou emissor para vinculá-lo aos seus
                      cartões.
                    </p>
                    <form
                      className="resource-form"
                      onSubmit={submitInstitution}
                    >
                      <Input
                        name="name"
                        label="Nome da instituição"
                        placeholder="Ex.: Nubank"
                        required
                        maxLength={120}
                      />
                      <Button type="submit" disabled={institutionPending}>
                        {institutionPending ? (
                          <Spinner size="sm" />
                        ) : (
                          <Save size={15} />
                        )}{" "}
                        {institutionPending
                          ? "Cadastrando..."
                          : "Salvar instituição"}
                      </Button>
                    </form>
                    <FormFeedback
                      pending={institutionPending}
                      pendingLabel="Salvando instituição..."
                      feedback={
                        institutionFeedback?.type === "error"
                          ? institutionFeedback
                          : null
                      }
                    />
                  </>
                ) : (
                  <>
                    {activeInstitutions.length === 0 ? (
                      <>
                        <p className="section-description">
                          Cadastre uma instituição antes de adicionar um cartão.
                        </p>
                        <Button
                          type="button"
                          onClick={() => setDrawer("institution")}
                        >
                          <Landmark size={16} /> Nova instituição
                        </Button>
                      </>
                    ) : (
                      <form className="resource-form" onSubmit={submitCard}>
                        <Input
                          name="name"
                          label="Nome do cartão"
                          placeholder="Ex.: Principal"
                          required
                          maxLength={120}
                        />
                        <label className="native-field">
                          <span>Instituição</span>
                          <select
                            name="institution_id"
                            required
                            defaultValue=""
                          >
                            <option value="" disabled>
                              Selecione
                            </option>
                            {activeInstitutions.map((institution) => (
                              <option
                                value={institution.id}
                                key={institution.id}
                              >
                                {institution.name}
                              </option>
                            ))}
                          </select>
                        </label>
                        <div className="form-grid-two">
                          <Input
                            name="effective_from"
                            type="month"
                            label="Configuração desde"
                            defaultValue={new Date().toISOString().slice(0, 7)}
                            required
                          />
                          <Input name="end_month" type="month" label="Até" />
                        </div>
                        <div className="form-grid-two">
                          <Input
                            name="nominal_due_day"
                            type="number"
                            min="1"
                            max="31"
                            label="Dia de vencimento"
                            defaultValue="6"
                            required
                          />
                          <Input
                            name="payment_month_offset"
                            type="number"
                            min="0"
                            max="12"
                            label="Offset de pagamento"
                            defaultValue="1"
                            required
                          />
                        </div>
                        <Input
                          name="context"
                          label="Contexto"
                          placeholder="Opcional"
                        />
                        <Button type="submit" disabled={cardPending}>
                          {cardPending ? (
                            <Spinner size="sm" />
                          ) : (
                            <Save size={15} />
                          )}{" "}
                          {cardPending ? "Cadastrando..." : "Salvar cartão"}
                        </Button>
                      </form>
                    )}
                    <FormFeedback
                      pending={cardPending}
                      pendingLabel="Salvando cartão..."
                      feedback={
                        cardFeedback?.type === "error" ? cardFeedback : null
                      }
                    />
                  </>
                )}
              </div>
            </>
          )}
        </DrawerContent>
      </Drawer>

      <section className="resource-list-section">
        <div className="section-heading compact">
          <div>
            <span className="section-kicker">MEUS CARTÕES</span>
            <h2>Cartões cadastrados</h2>
          </div>
          <ReceiptText size={20} />
        </div>
        {actionError && <Alert color="error">{actionError}</Alert>}
        {cards.length === 0 ? (
          <div className="empty-resource-inline">Nenhum cartão cadastrado.</div>
        ) : (
          <div className="credit-card-list">
            {cards.map((card) =>
              editing === card.id ? (
                <CardEditor
                  card={card}
                  key={card.id}
                  loading={pendingAction === card.id}
                  onCancel={() => setEditing(null)}
                  onSave={(name) =>
                    action(card.id, `/credit-cards/${card.id}`, "PATCH", {
                      name,
                    })
                  }
                />
              ) : (
                <CreditCardTile
                  key={card.id}
                  card={card}
                  loading={pendingAction === card.id}
                  onEdit={() => setEditing(card.id)}
                  onArchive={() =>
                    action(
                      card.id,
                      `/credit-cards/${card.id}/archive`,
                      "POST",
                      {},
                    )
                  }
                />
              ),
            )}
          </div>
        )}
      </section>
    </div>
  );
}

function FormFeedback({
  pending,
  pendingLabel,
  feedback,
}: {
  pending: boolean;
  pendingLabel: string;
  feedback: Feedback;
}) {
  return (
    <div aria-live="polite" aria-atomic="true">
      {pending ? (
        <div className="form-feedback loading">
          <Spinner size="sm" />
          <span>{pendingLabel}</span>
        </div>
      ) : feedback?.type === "success" ? (
        <div className="form-feedback success" role="status">
          <Check size={17} />
          <span>{feedback.message}</span>
        </div>
      ) : feedback?.type === "error" ? (
        <Alert color="error">{feedback.message}</Alert>
      ) : null}
    </div>
  );
}

function CreditCardTile({
  card,
  loading,
  onEdit,
  onArchive,
}: {
  card: CreditCard;
  loading: boolean;
  onEdit: () => void;
  onArchive: () => void;
}) {
  const configuration = card.configurations.at(-1);
  const theme = institutionTheme(card.institution.name);
  return (
    <article className="credit-card-tile">
      <div className={`credit-card-face brand-${theme}`}>
        <div className="credit-card-face-top">
          <span className="credit-card-institution">
            {card.institution.name}
          </span>
          <span className="credit-card-state">
            {card.status === "active" ? "Ativo" : "Arquivado"}
          </span>
        </div>
        <div className="credit-card-symbols">
          <span className="credit-card-chip" aria-hidden="true">
            <i />
            <i />
          </span>
          <Wifi size={20} aria-hidden="true" />
        </div>
        <div className="credit-card-face-bottom">
          <div>
            <small>CARTÃO DE CRÉDITO</small>
            <strong>{card.name}</strong>
          </div>
          <CreditCardIcon size={24} aria-hidden="true" />
        </div>
      </div>
      <div className="credit-card-info">
        <span>
          <CalendarDays size={15} /> Vence dia{" "}
          {configuration?.nominal_due_day ?? "—"}
        </span>
        <span>
          Pagamento:{" "}
          {configuration?.payment_month_offset === 0
            ? "mesmo mês"
            : configuration?.payment_month_offset === 1
              ? "mês seguinte"
              : `${configuration?.payment_month_offset ?? "—"} meses depois`}
        </span>
      </div>
      <div className="credit-card-actions">
        <Link href={`/invoices/${card.id}`}>
          Ver faturas <ArrowRight size={16} />
        </Link>
        {card.status === "active" && (
          <div>
            <Button
              variant="ghost"
              size="sm"
              onClick={onEdit}
              disabled={loading}
            >
              <Pencil size={15} /> Editar
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={onArchive}
              disabled={loading}
            >
              {loading ? <Spinner size="sm" /> : <Archive size={15} />} Arquivar
            </Button>
          </div>
        )}
      </div>
    </article>
  );
}

function CardEditor({
  card,
  loading,
  onCancel,
  onSave,
}: {
  card: CreditCard;
  loading: boolean;
  onCancel: () => void;
  onSave: (name: string) => void;
}) {
  return (
    <Card className="resource-card" padding="none">
      <CardContent className="resource-card-content">
        <form
          className="resource-form"
          onSubmit={(event) => {
            event.preventDefault();
            onSave(
              String(new FormData(event.currentTarget).get("name")).trim(),
            );
          }}
        >
          <Input
            name="name"
            label="Nome do cartão"
            defaultValue={card.name}
            required
          />
          <div className="form-actions">
            <Button type="submit" disabled={loading}>
              {loading ? <Spinner size="sm" /> : <Save size={15} />} Salvar
            </Button>
            <Button type="button" variant="ghost" onClick={onCancel}>
              Cancelar
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
