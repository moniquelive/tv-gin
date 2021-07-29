module Main exposing (..)

import Array exposing (Array)
import Browser
import Dict exposing (Dict)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import Url exposing (percentEncode)



-- MAIN


main =
    Browser.element
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }



-- MODEL


type alias Meme =
    { id : String
    , name : String
    , filename : String
    , font_color : String
    , line_chars : Int
    , boxes : List (List Int)
    }


defaultMeme =
    Meme "" "" "" "" 0 []


type State
    = Failure
    | Loading
    | Success


-- TODO: remover selectedMeme se for desnecessario
type alias Model =
    { memes : Dict String Meme
    , selectedMeme : String
    , texts : Array String
    , state : State
    }


defaultModel =
    Model Dict.empty "" Array.empty Loading


init : () -> ( Model, Cmd Msg )
init _ =
    ( defaultModel, getConfig )



-- UPDATE


type Msg
    = UpdateURL
    | RefreshConfig
    | GotConfig (Result Http.Error (Dict String Meme))
    | UpdateMeme String
    | TextChanged Int String


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        UpdateURL ->
            ( model, getConfig )

        RefreshConfig ->
            ( defaultModel, getConfig )

        UpdateMeme id ->
            let
                meme =
                    Dict.get id model.memes |> Maybe.withDefault defaultMeme

                numTexts =
                    List.length meme.boxes

                previousText n =
                    Array.get n model.texts |> Maybe.withDefault ""
            in
            ( { model | selectedMeme = id, texts = Array.initialize numTexts previousText }, Cmd.none )

        GotConfig result ->
            case result of
                Ok memes ->
                    let
                        firstMeme =
                            List.head (Dict.values memes) |> Maybe.withDefault defaultMeme

                        numTexts =
                            List.length firstMeme.boxes
                    in
                    ( Model memes firstMeme.id (Array.repeat numTexts "") Success, Cmd.none )

                Err _ ->
                    ( { model | state = Failure }, Cmd.none )

        TextChanged index str ->
            ( { model | texts = Array.set index str model.texts }, Cmd.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Model -> Html Msg
view model =
    case model.state of
        Failure ->
            div []
                [ text "Não consegui carregar a configuração inicial. Tente novamente. "
                , button [ onClick RefreshConfig ] [ text "De novo!" ]
                ]

        Loading ->
            text "Carregando..."

        Success ->
            div []
                [ h1 [] [ text "Escolha o Meme:" ]
                , div [ class "row" ]
                    [ div [ class "col-md-4 mb-4 mb-md-0" ]
                        [ select [ value model.selectedMeme, class "mb-4 form-control", onInput UpdateMeme ]
                            [ optgroup [ attribute "label" "Implementados" ]
                                (List.map viewMeme (Dict.values model.memes))
                            ]
                        , fieldset [] (Array.toIndexedList model.texts |> List.map createText)
                        ]
                    , div [ class "col-md-8" ]
                        [ img
                            [ class "mw-100"
                            , src ("/meme?meme=" ++ model.selectedMeme ++ Array.foldl createTextParams "&" model.texts)
                            ]
                            []
                        ]
                    ]
                ]


viewMeme : Meme -> Html Msg
viewMeme meme =
    option [ value meme.id ] [ text meme.name ]


createText : ( Int, String ) -> Html Msg
createText ( index, str ) =
    input
        [ onInput (TextChanged index)
        , class "form-control"
        , type_ "text"
        , value str
        ]
        []


createTextParams : String -> String -> String
createTextParams str acc =
    acc ++ percentEncode "text[]" ++ "=" ++ percentEncode str ++ "&"



-- HTTP


getConfig : Cmd Msg
getConfig =
    Http.get
        { url = "/config.json"
        , expect = Http.expectJson GotConfig configDecoder
        }



-- JSON decode


configDecoder : D.Decoder (Dict String Meme)
configDecoder =
    D.dict memeDecoder


memeDecoder : D.Decoder Meme
memeDecoder =
    D.map6 Meme
        (D.field "id" D.string)
        (D.field "name" D.string)
        (D.field "filename" D.string)
        (D.field "font-color" D.string)
        (D.field "line-chars" D.int)
        (D.field "boxes" (D.list (D.list D.int)))
